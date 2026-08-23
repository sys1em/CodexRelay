/*
 * @Author        : 顾青离
 * @Url           : sucaijun.com
 * @Email         : Ricky@LiHai.La
 * @Project       : CodexRelay
 * @Description   : Codex API 中转热切换桌面工具
 * @File          : 不改变响应字节的用量旁路观察器
 */
package usage

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strings"
)

const maxObservedBytes = 8 << 20

type Observed struct {
	Status       string
	Model        string
	InputTokens  uint64
	OutputTokens uint64
	CachedTokens uint64
	TotalTokens  uint64
}

type Observer struct {
	mode      string
	pending   []byte
	eventName string
	eventData bytes.Buffer
	jsonData  bytes.Buffer
	dropping  bool
	finished  bool
	result    Observed
}

type observedReadCloser struct {
	io.ReadCloser
	observer *Observer
}

func (r *observedReadCloser) Read(buffer []byte) (int, error) {
	count, err := r.ReadCloser.Read(buffer)
	if count > 0 {
		r.observer.Observe(buffer[:count])
	}
	return count, err
}

func NewObserver(header http.Header) *Observer {
	observer := &Observer{result: Observed{Status: StatusUnreported}}
	encoding := strings.TrimSpace(strings.ToLower(header.Get("Content-Encoding")))
	if encoding != "" && encoding != "identity" {
		return observer
	}
	contentType, _, _ := mime.ParseMediaType(header.Get("Content-Type"))
	switch strings.ToLower(contentType) {
	case "text/event-stream":
		observer.mode = "sse"
	case "application/json", "application/problem+json":
		observer.mode = "json"
	}
	return observer
}

func (o *Observer) Enabled() bool {
	return o != nil && o.mode != ""
}

func (o *Observer) Wrap(body io.ReadCloser) io.ReadCloser {
	return &observedReadCloser{ReadCloser: body, observer: o}
}

func (o *Observer) Observe(chunk []byte) {
	if !o.Enabled() || o.finished || len(chunk) == 0 {
		return
	}
	if o.mode == "json" {
		if o.jsonData.Len()+len(chunk) > maxObservedBytes {
			o.mode = ""
			o.jsonData.Reset()
			return
		}
		_, _ = o.jsonData.Write(chunk)
		return
	}
	o.pending = append(o.pending, chunk...)
	for {
		newline := bytes.IndexByte(o.pending, '\n')
		if newline < 0 {
			if len(o.pending) > maxObservedBytes {
				o.dropping = true
				o.pending = o.pending[:0]
			}
			return
		}
		line := o.pending[:newline]
		o.pending = o.pending[newline+1:]
		line = bytes.TrimSuffix(line, []byte{'\r'})
		o.processSSELine(line)
		if o.finished {
			o.pending = nil
			return
		}
	}
}

func (o *Observer) processSSELine(line []byte) {
	if len(line) == 0 {
		o.finishSSEEvent()
		return
	}
	if o.dropping {
		return
	}
	if bytes.HasPrefix(line, []byte("event:")) {
		o.eventName = strings.TrimSpace(string(line[len("event:"):]))
		return
	}
	if !bytes.HasPrefix(line, []byte("data:")) {
		return
	}
	data := bytes.TrimPrefix(line, []byte("data:"))
	data = bytes.TrimPrefix(data, []byte(" "))
	if o.eventData.Len() > 0 {
		_ = o.eventData.WriteByte('\n')
	}
	if o.eventData.Len()+len(data) > maxObservedBytes {
		o.dropping = true
		o.eventData.Reset()
		return
	}
	_, _ = o.eventData.Write(data)
}

func (o *Observer) finishSSEEvent() {
	defer func() {
		o.eventName = ""
		o.eventData.Reset()
		o.dropping = false
	}()
	if o.dropping || o.eventData.Len() == 0 {
		return
	}
	data := o.eventData.Bytes()
	if o.eventName != "response.completed" && !bytes.Contains(data, []byte(`"response.completed"`)) {
		return
	}
	if observed, ok := parseObserved(data, true); ok {
		o.result = observed
		o.finished = true
	}
}

func (o *Observer) Finish() Observed {
	if o == nil {
		return Observed{Status: StatusUnreported}
	}
	if o.finished {
		return o.result
	}
	if o.mode == "sse" {
		if len(o.pending) > 0 {
			line := bytes.TrimSuffix(o.pending, []byte{'\r'})
			o.processSSELine(line)
		}
		o.finishSSEEvent()
	} else if o.mode == "json" && o.jsonData.Len() > 0 {
		if observed, ok := parseObserved(o.jsonData.Bytes(), false); ok {
			o.result = observed
		}
	}
	o.finished = true
	return o.result
}

func parseObserved(data []byte, expectEnvelope bool) (Observed, bool) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var root map[string]any
	if err := decoder.Decode(&root); err != nil {
		return Observed{}, false
	}
	body := root
	if expectEnvelope {
		if eventType, _ := root["type"].(string); eventType != "response.completed" {
			return Observed{}, false
		}
		response, ok := root["response"].(map[string]any)
		if !ok {
			return Observed{}, false
		}
		body = response
	}
	usageData, ok := body["usage"].(map[string]any)
	if !ok {
		return Observed{}, false
	}
	input, inputOK := unsignedJSONNumber(usageData["input_tokens"])
	output, outputOK := unsignedJSONNumber(usageData["output_tokens"])
	if !inputOK || !outputOK {
		return Observed{}, false
	}
	cached := uint64(0)
	if details, ok := usageData["input_tokens_details"].(map[string]any); ok {
		cached, _ = unsignedJSONNumber(details["cached_tokens"])
	}
	total, ok := unsignedJSONNumber(usageData["total_tokens"])
	if !ok {
		total = input + output
	}
	model, _ := body["model"].(string)
	return Observed{
		Status: StatusReported, Model: model, InputTokens: input,
		OutputTokens: output, CachedTokens: cached, TotalTokens: total,
	}, true
}

func unsignedJSONNumber(value any) (uint64, bool) {
	switch number := value.(type) {
	case json.Number:
		parsed, err := number.Int64()
		return uint64(parsed), err == nil && parsed >= 0
	case float64:
		if number < 0 || number != float64(uint64(number)) {
			return 0, false
		}
		return uint64(number), true
	default:
		return 0, false
	}
}
