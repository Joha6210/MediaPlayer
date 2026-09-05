package player

import "sync"

type MockPlayer struct {
	mu      sync.RWMutex
	state   MockState
	playing bool
}

type MockState struct {
	URL     string            `json:"url"`
	Volume  int               `json:"volume"`
	Paused  bool              `json:"paused"`
	Headers map[string]string `json:"headers,omitempty"`
}

func NewMockPlayer(defaultVolume int) *MockPlayer {
	return &MockPlayer{
		state: MockState{
			Volume: defaultVolume,
		},
	}
}

func (m *MockPlayer) Play(url string, volume int, headers map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state.URL = url
	m.state.Volume = volume
	m.state.Paused = false
	m.state.Headers = cloneHeaders(headers)
	m.playing = true
	return nil
}

func (m *MockPlayer) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.playing = false
	m.state.Paused = false
	return nil
}

func (m *MockPlayer) Pause(paused bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state.Paused = paused
	return nil
}

func (m *MockPlayer) SetVolume(volume int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state.Volume = volume
	return nil
}

func (m *MockPlayer) Snapshot() (MockState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return MockState{
		URL:     m.state.URL,
		Volume:  m.state.Volume,
		Paused:  m.state.Paused,
		Headers: cloneHeaders(m.state.Headers),
	}, m.playing
}

func (m *MockPlayer) ListenEvents() (<-chan struct {
	Title  string
	Paused bool
}, error) {
	// For the mock player, we can return a closed channel since it doesn't generate events.
	ch := make(chan struct {
		Title  string
		Paused bool
	})
	close(ch)
	return ch, nil
}

func cloneHeaders(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return nil
	}
	dup := make(map[string]string, len(headers))
	for k, v := range headers {
		dup[k] = v
	}
	return dup
}
