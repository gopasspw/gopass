package pinentry

import "fmt"

func ExampleClient() {
	pi, err := New()
	if err != nil {
		panic(err)
	}
	_ = pi.Set("title", "Agent Pinentry")
	_ = pi.Set("desc", "Asking for a passphrase")
	_ = pi.Set("prompt", "Please enter your passphrase:")
	_ = pi.Set("ok", "OK")
	pin, err := pi.GetPin()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(pin))
}
