package audit

import "encoding/json"

func Encode(e Event) ([]byte, error) { return json.Marshal(e) }
func Decode(b []byte) (Event, error) { var e Event; err := json.Unmarshal(b, &e); return e, err }
