package ollama

// // // GenerateRequest represents the request body sent to the Ollama API.
// // // Adjust fields to match your actual Ollama API requirements.
// // type GenerateRequest struct {
// // 	Prompt      string   `json:"prompt"`
// // 	Model       string   `json:"model,omitempty"`
// // 	Stream      *bool    `json:"stream,omitempty"`
// // 	Temperature *float64 `json:"temperature,omitempty"`
// // 	// Add additional fields needed by your API
// // }

// // // GenerateResponse represents the response returned from the Ollama API.
// // // Adjust fields to match your actual Ollama API responses.
// // type GenerateResponse struct {
// // 	Response string `json:"response"`
// // 	// Add additional fields needed by your API
// // }

// // OllamaClient is an interface that defines how to communicate with the Ollama API.
// type OllamaClient interface {
// 	// Generate sends a prompt to Ollama and returns the generated text.
// 	Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error)
// }

// type client struct {
// 	httpClient *http.Client
// 	baseURL    string
// 	// you might also store any necessary tokens, headers, etc.
// }

// // Option is a functional option for configuring the client.
// type Option func(*client)

// // WithHTTPClient sets a custom http.Client.
// func WithHTTPClient(hc *http.Client) Option {
// 	return func(c *client) {
// 		c.httpClient = hc
// 	}
// }

// // WithBaseURL sets a custom base URL for the Ollama API.
// func WithBaseURL(url string) Option {
// 	return func(c *client) {
// 		c.baseURL = url
// 	}
// }

// // NewClient constructs a new client with optional arguments.
// func NewClient(opts ...Option) OllamaClient {
// 	c := &client{
// 		httpClient: &http.Client{
// 			Timeout: 15 * time.Second,
// 		},
// 		baseURL: "http://localhost:11434", // Default local Ollama endpoint
// 	}
// 	for _, opt := range opts {
// 		opt(c)
// 	}
// 	return c
// }

// // Generate implements the OllamaClient interface, by sending a prompt to Ollama and returning a response.
// func (c *client) Generate(ctx context.Context, request GenerateRequest) (*GenerateResponse, error) {

// 	stream := false
// 	if request.Stream == nil {
// 		request.Stream = &stream
// 	}
// 	url := fmt.Sprintf("%s/api/generate", c.baseURL)
// 	// Marshal request
// 	b, err := json.Marshal(request)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to marshal request: %w", err)
// 	}

// 	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(b))
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create http request: %w", err)
// 	}

// 	// // Prepare the HTTP request
// 	// httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
// 	// if err != nil {
// 	// 	return nil, fmt.Errorf("failed to create http request: %w", err)
// 	// }
// 	httpReq.Header.Set("Content-Type", "application/json")
// 	// If you need an authorization header, do something like:
// 	// httpReq.Header.Set("Authorization", "Bearer <TOKEN>")

// 	// Write the marshaled request to the request body
// 	// httpReq.Body = makeNopCloser(requestBody)

// 	// Execute the request
// 	resp, err := c.httpClient.Do(httpReq)
// 	if err != nil {
// 		return nil, fmt.Errorf("http do request failed: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("generate request failed with status code %d", resp.StatusCode)
// 	}

// 	// Unmarshal the response
// 	var generateResp GenerateResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&generateResp); err != nil {
// 		return nil, fmt.Errorf("failed to decode response: %w", err)
// 	}

// 	return &generateResp, nil
// }

// // makeNopCloser helps convert a byte slice into an io.ReadCloser.
// func makeNopCloser(data []byte) *nopCloser {
// 	return &nopCloser{data: data}
// }

// type nopCloser struct {
// 	data   []byte
// 	read   int
// 	closed bool
// }

// func (c *nopCloser) Read(p []byte) (int, error) {
// 	if c.closed {
// 		return 0, errors.New("nopCloser: read after close")
// 	}
// 	if c.read >= len(c.data) {
// 		return 0, nil
// 	}
// 	n := copy(p, c.data[c.read:])
// 	c.read += n
// 	return n, nil
// }

// func (c *nopCloser) Close() error {
// 	c.closed = true
// 	return nil
// }
