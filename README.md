# websocket-server

A Golang WebSocket server designed to handle multiple concurrent connections.

## Steps to install and run the server

1. Clone the repository: git clone https://github.com/your-repo/websocket-server-go.git
2. Change directory to 'src' from root of the project: cd src
3. Run the main file: go run main.go
4. Go to localhost:8080/ to check if the server is up successully. You will see a message saying 'This is home!'
5. The websocket server is on path '/pingpong' so every request to localhost:8080/pingpong will be upgraded to websocket connection

## Steps to run the tests

1. Change directory to 'test' from root of the project: cd test
2. Run the main_test file: go test main_test.go -v (-v is important as it allows to see what actually is happening at the server side)


## Future enhancements
1. Support multiple groups
2. Give application constants via command line on startup or introduce a config file 
3. Better error handling 
4. Stress testing to check how many websockets can be handled concurrently without affecting the performance too much

## Wierd Things

1. https://github.com/gorilla/websocket/issues/744 - ReadMessage doesn't forward control messages

## Troubleshooting 

1. Please feel free to contact me in case of any queries
