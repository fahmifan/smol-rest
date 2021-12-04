package definitions

// GreeterService makes nice greetings.
type GreeterService interface {
	// Greet makes a greeting.
	Greet(GreetRequest) GreetResponse
}

// GreetRequest is the request object for GreeterService.Greet.
type GreetRequest struct {
	// Name is the person to greet.
	// example: "Mat Ryer"
	Name string
}

// GreetResponse is the response object containing a
// person's greeting.
type GreetResponse struct {
	// Greeting is the greeting that was generated.
	// example: "Hello Mat Ryer"
	Greeting string
}
