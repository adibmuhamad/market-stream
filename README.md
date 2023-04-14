## Real-time Stock Price WebSocket Server

This project is a WebSocket server that provides real-time stock price. Clients can subscribe to updates for specific stock symbols by connecting to the server and providing the symbol as a query parameter.

## Features

1. Real-time stock price updates (Open, High, Low, Close, Previous Close, Timestamp, Volume, Value, Change, and Change Percent)
2. Pub/Sub model for efficient data distribution

## Usage

1. Ensure you have Go installed on your system.
2. Clone the repository and navigate to the project directory.
3. Run the server with `go run main.go`.
4. Connect to the websocket server with a client using a URL in the following format: `ws://localhost:8080/stock?symbol=<SYMBOLS>`.

- Replace `<SYMBOLS>` with stock symbols (e.g., AAPL).

## Contributing

Feel free to contribute! Here's how you can contribute:

- [Open an issue](https://github.com/adibmuhamad/market-stream/issues) if you believe you've encountered a bug.
- Make a [pull request](https://github.com/adibmuhamad/market-stream/pull) to add new features/make quality-of-life improvements/fix bugs.

## Author

- Muhammad Adib Yusrul Muna

## License
Copyright © 2023 Muhammad Adib Yusrul Muna

This software is distributed under the MIT license. See the [LICENSE](https://github.com/adibmuhamad/market-stream/blob/main/LICENSE) file for the full license text.