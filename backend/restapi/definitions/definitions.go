package definitions

// SmolService makes nice Loginings.
type SmolService interface {
	// Login makes a Logining.
	Login(LoginRequest) LoginResponse
}

// LoginRequest is the request object for SmolService.Login.
type LoginRequest struct {
	// Name is the person to Login.
	// example: "Mat Ryer"
	Name string
}

// LoginResponse is the response object containing a
// person's Logining.
type LoginResponse struct {
	// Logining is the Logining that was generated.
	// example: "Hello Mat Ryer"
	Logining string
}
