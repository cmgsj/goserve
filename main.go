package main

import "goserve/cmd"

func main() {
	cmd.Execute()
}

// var (
// 	port = flag.Int("port", 1234, "port to listen on")
// 	path = flag.String("path", ".", "path to serve")
// )

// func main() {
// 	flag.Parse()
// 	log.Printf("serving %s at http://localhost:%d\n", *path, *port)
// 	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), http.FileServer(http.Dir(*path))))
// }
