package flow

import "encoding/json"

func Parse(b []byte) (*Flow, error) {
	var steps []Step
	if err := json.Unmarshal(b, &steps); err != nil {
		return nil, err
	}

	m := make(map[int]Step, len(steps))
	start := 0
	for _, st := range steps {
		if start == 0 || st.Sequencia < start {
			start = st.Sequencia
		}
		m[st.Sequencia] = st
	}

	return &Flow{
		StartSeq: start,
		Steps:    m,
	}, nil
}
