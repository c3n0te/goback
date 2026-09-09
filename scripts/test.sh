curl.exe -X POST http://127.0.0.1:8000/account -H "Content-Type: application/json" -d '{\"Username\":\"c3n0te\", \"Email\":\"c3n0te@gmail.com\", \"Password\":\"secret\", \"Balance\":1000.0}'
curl.exe -X POST http://127.0.0.1:8000/account -H "Content-Type: application/json" -d '{\"Username\":\"t0rch\", \"Email\":\"t0rch@gmail.com\", \"Password\":\"secret\", \"Balance\":5000.0}'
curl.exe -X POST http://127.0.0.1:8000/buy -H "Content-Type: application/json" -d '{\"Type\":\"Buy\", \"UserId\":\"<Fill This In>\", \"Stock\":\"AAPL\", \"Shares\":5000.0}'
curl.exe -X POST http://127.0.0.1:8000/sell -H "Content-Type: application/json" -d '{\"Type\":\"Sell\", \"UserId\":\"<Fill This In>\", \"Stock\":\"TSLA\", \"Shares\":500.0}'
curl.exe http://127.0.0.1:8000/transactions/<FILL THIS IN>
