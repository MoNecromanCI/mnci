package plugingo

// Greet greets someone by name: a worked example of a use case. Replace it with your own.
func Greet(name string) Greeting {
	return Greeting{Message: "Hello, " + name + "!"}
}
