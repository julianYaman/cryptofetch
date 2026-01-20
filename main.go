package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const version = "1.0.0"

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorBold   = "\033[1m"
	colorOrange = "\033[38;5;208m"
	colorPink   = "\033[38;5;205m"
)

// Cryptocurrency colors
var coinColors = map[string]string{
	"bitcoin":  colorOrange,
	"ethereum": "\033[38;5;99m", // Purple/Blue
	"cardano":  colorBlue,
	"solana":   "\033[38;5;141m", // Purple
	"ripple":   "\033[38;5;33m",  // Blue
	"dogecoin": colorYellow,
	"polkadot": colorPink,
	"default":  colorCyan,
}

type CoinData struct {
	ID                string  `json:"id"`
	Symbol            string  `json:"symbol"`
	Name              string  `json:"name"`
	CurrentPrice      float64 `json:"current_price"`
	MarketCap         float64 `json:"market_cap"`
	MarketCapRank     int     `json:"market_cap_rank"`
	TotalVolume       float64 `json:"total_volume"`
	High24h           float64 `json:"high_24h"`
	Low24h            float64 `json:"low_24h"`
	PriceChange1h     float64 `json:"price_change_percentage_1h_in_currency"`
	PriceChange24h    float64 `json:"price_change_percentage_24h"`
	PriceChange7d     float64 `json:"price_change_percentage_7d_in_currency"`
	CirculatingSupply float64 `json:"circulating_supply"`
	TotalSupply       float64 `json:"total_supply"`
	ATH               float64 `json:"ath"`
	ATHDate           string  `json:"ath_date"`
}

// Generate ASCII art on "https://emojicombos.com/dot-art-generator"
// With threshold=0.50, width=30, characters=464
var asciiArt = map[string]string{
	"bitcoin": `
⠀⠀⠀⠀⠀⠀⠀⠀⣀⣤⣴⣶⣾⣿⣿⣿⣿⣷⣶⣦⣤⣀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⣠⣴⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⣄⠀⠀⠀⠀⠀
⠀⠀⠀⣠⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣄⠀⠀⠀
⠀⠀⣴⣿⣿⣿⣿⣿⣿⣿⠟⠿⠿⡿⠀⢰⣿⠁⢈⣿⣿⣿⣿⣿⣿⣿⣿⣦⠀⠀
⠀⣼⣿⣿⣿⣿⣿⣿⣿⣿⣤⣄⠀⠀⠀⠈⠉⠀⠸⠿⣿⣿⣿⣿⣿⣿⣿⣿⣧⠀
⢰⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡏⠀⠀⢠⣶⣶⣤⡀⠀⠈⢻⣿⣿⣿⣿⣿⣿⣿⡆
⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠃⠀⠀⠼⣿⣿⡿⠃⠀⠀⢸⣿⣿⣿⣿⣿⣿⣿⣷
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡟⠀⠀⢀⣀⣀⠀⠀⠀⠀⢴⣿⣿⣿⣿⣿⣿⣿⣿⣿
⢿⣿⣿⣿⣿⣿⣿⣿⢿⣿⠁⠀⠀⣼⣿⣿⣿⣦⠀⠀⠈⢻⣿⣿⣿⣿⣿⣿⣿⡿
⠸⣿⣿⣿⣿⣿⣿⣏⠀⠀⠀⠀⠀⠛⠛⠿⠟⠋⠀⠀⠀⣾⣿⣿⣿⣿⣿⣿⣿⠇
⠀⢻⣿⣿⣿⣿⣿⣿⣿⣿⠇⠀⣤⡄⠀⣀⣀⣀⣀⣠⣾⣿⣿⣿⣿⣿⣿⣿⡟⠀
⠀⠀⠻⣿⣿⣿⣿⣿⣿⣿⣄⣰⣿⠁⢀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⠀⠀
⠀⠀⠀⠙⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠋⠀⠀⠀
⠀⠀⠀⠀⠀⠙⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⠋⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠉⠛⠻⠿⢿⣿⣿⣿⣿⡿⠿⠟⠛⠉⠀⠀⠀⠀⠀⠀⠀⠀`,

	"ethereum": `
⠀⠀⠀⠀⠀⠀⠀⠀⣀⣤⣴⣶⣾⣿⣿⣿⣿⣷⣶⣦⣤⣀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⣠⣴⣿⣿⣿⣿⣿⣿⣿⣿⢿⣿⣿⣿⣿⣿⣿⣿⣦⣄⠀⠀⠀⠀⠀
⠀⠀⠀⣠⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⠏⠀⠹⣿⣿⣿⣿⣿⣿⣿⣿⣷⣄⠀⠀⠀
⠀⠀⣴⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠃⠀⠀⠀⠙⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⠀⠀
⠀⣼⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠁⠀⠀⠀⠀⠀⠈⢿⣿⣿⣿⣿⣿⣿⣿⣿⣧⠀
⢰⣿⣿⣿⣿⣿⣿⣿⣿⣿⡟⠀⠀⠀⠀⣀⠀⠀⠀⠈⢻⣿⣿⣿⣿⣿⣿⣿⣿⡆
⣾⣿⣿⣿⣿⣿⣿⣿⣿⠏⠀⣀⡤⠖⠛⠉⠛⠶⣤⣀⠀⠹⣿⣿⣿⣿⣿⣿⣿⣷
⣿⣿⣿⣿⣿⣿⣿⣿⡿⠞⠋⠁⠀⠀⠀⠀⠀⠀⠀⠈⠙⠳⣿⣿⣿⣿⣿⣿⣿⣿
⢿⣿⣿⣿⣿⣿⣿⣿⣿⡳⢦⣄⠀⠀⠀⠀⠀⠀⠀⣠⡴⢚⣿⣿⣿⣿⣿⣿⣿⡿
⠸⣿⣿⣿⣿⣿⣿⣿⣿⣷⣄⠈⠙⠶⣄⣀⣤⠖⠋⠁⣠⣿⣿⣿⣿⣿⣿⣿⣿⠇
⠀⢻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⡀⠀⠀⠉⠀⠀⢀⣴⣿⣿⣿⣿⣿⣿⣿⣿⡟⠀
⠀⠀⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⡄⠀⠀⠀⣠⣾⣿⣿⣿⣿⣿⣿⣿⣿⠟⠀⠀
⠀⠀⠀⠙⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣆⠀⣴⣿⣿⣿⣿⣿⣿⣿⣿⡿⠋⠀⠀⠀
⠀⠀⠀⠀⠀⠙⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⠋⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠉⠛⠻⠿⢿⣿⣿⣿⣿⡿⠿⠟⠛⠉⠀⠀⠀⠀⠀⠀⠀⠀`,

	"cardano": `
⠀⠀⠀⠀⠀⠀⠀⣀⣤⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣤⣀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⢀⣴⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⡀⠀⠀⠀⠀
⠀⠀⢀⣴⣿⣿⣿⣿⣿⡟⢿⣿⣿⣿⣏⣹⣿⣿⣿⣿⠛⣿⣿⣿⣿⣿⣦⡀⠀⠀
⠀⢠⣾⣿⣿⣿⣿⣿⣿⣿⡿⠿⣿⣿⣿⣿⣿⣿⠿⢿⣿⣿⣿⣿⣿⣿⣿⣿⡄⠀
⢠⣿⣿⣿⣿⣿⠿⣿⣿⣿⣷⣴⣿⣿⡅⢀⣿⣿⣧⣼⣿⣿⣿⠿⣿⣿⣿⣿⣿⡄
⣿⣿⣿⣿⣿⣿⣦⣿⣿⠛⠻⣿⠋⠈⠹⡟⠉⠉⢻⡿⠛⢿⣿⣦⣿⣿⣿⣿⣿⡟
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⡾⠿⣤⣠⣼⣷⣤⣤⠾⢷⣶⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣩⣿⣏⣈⣿⣿⡁⠀⢈⣿⣿⣿⣿⡇⠀⠀⣿⣿⣇⣈⣿⣏⣹⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⠷⣶⠚⠙⢻⡟⠋⠙⢶⡾⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⠟⣿⣿⣤⣶⣿⣄⣀⣰⣧⣀⣀⣾⣷⣤⣾⣿⠟⣿⣿⣿⣿⣿⠇
⠘⣿⣿⣿⣿⣿⣶⣿⣿⣿⡿⠻⣿⣿⡃⠈⣿⣿⡟⢻⣿⣿⣿⣶⣿⣿⣿⣿⣿⠀
⠀⠘⢿⣿⣿⣿⣿⣿⣿⣿⣷⣶⣿⣿⣿⣿⣿⣿⣷⣾⣿⣿⣿⣿⣿⣿⣿⣿⠃⠀
⠀⠀⠈⠻⣿⣿⣿⣿⣿⣧⣾⣿⣿⣿⣏⣹⣿⣿⣿⣿⣤⣿⣿⣿⣿⣿⠟⠁⠀⠀
⠀⠀⠀⠀⠈⠻⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠟⠁⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠉⠛⠿⢿⢿⣿⣿⣿⣿⣿⣿⡿⣿⠿⠛⠉⠀⠀⠀⠀⠀⠀⠀`,

	"solana": `
⠀⠀⠀⠀⠀⠀⠀⠀⣀⣤⣴⣶⣾⣿⣿⣿⣿⣷⣶⣦⣤⣀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⣠⣶⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣄⠀⠀⠀⠀⠀
⠀⠀⠀⣠⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣄⠀⠀⠀
⠀⠀⣼⣿⣿⣿⣿⣿⣿⠟⠛⠛⠛⠛⠛⠛⠛⠛⠛⠛⠛⠛⠛⣻⣿⣿⣿⣧⠀⠀
⠀⣼⣿⣿⣿⣿⣿⠟⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣴⣿⣿⣿⣿⣿⣧⠀
⢰⣿⣿⣿⣿⣿⣧⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣴⣿⣿⣿⣿⣿⣿⣿⣿⡆
⣾⣿⣿⣿⣿⣿⣟⠛⠛⠛⠛⠛⠛⠛⠛⠛⠛⠛⠛⠛⠻⣿⣿⣿⣿⣿⣿⣿⣿⣷
⣿⣿⣿⣿⣿⣿⣿⣦⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠻⣿⣿⣿⣿⣿⣿⣿
⢿⣿⣿⣿⣿⣿⣿⣿⣿⣦⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣼⣿⣿⣿⣿⣿⡿
⠸⣿⣿⣿⣿⣿⣿⣿⣿⠟⠛⠛⠛⠛⠛⠛⠛⠛⠛⠛⠛⠛⠛⣻⣿⣿⣿⣿⣿⠇
⠀⢻⣿⣿⣿⣿⣿⠟⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣼⣿⣿⣿⣿⣿⡟⠀
⠀⠀⢻⣿⣿⣿⣧⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣴⣿⣿⣿⣿⣿⣿⡟⠀⠀
⠀⠀⠀⠙⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠋⠀⠀⠀
⠀⠀⠀⠀⠀⠙⠿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠿⠋⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠉⠛⠻⠿⢿⣿⣿⣿⣿⡿⠿⠟⠛⠉⠀⠀⠀⠀⠀⠀⠀⠀`,

	"ripple": `
⠀⠀⠀⠀⠀⠀⠀⠀⣀⣤⣴⣶⣾⣿⣿⣿⣿⣿⣶⣶⣤⣀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⣠⣴⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣄⠀⠀⠀⠀⠀
⠀⠀⠀⣠⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣄⠀⠀⠀
⠀⠀⣴⣿⣿⣯⡉⠉⠙⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⠉⢉⣽⣿⣿⣧⡀⠀
⠀⣼⣿⣿⣿⣿⣿⣦⡀⠈⠻⣿⣿⣿⣿⣿⣿⣿⣿⠟⠁⢀⣴⣿⣿⣿⣿⣿⣷⠀
⢰⣿⣿⣿⣿⣿⣿⣿⣿⣦⡀⠈⠻⢿⣿⣿⣿⠟⠁⠀⣰⣿⣿⣿⣿⣿⣿⣿⣿⣇
⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⣀⠀⠀⠁⠀⢀⣤⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡷⢶⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⠉⠀⠀⠀⠀⠉⠛⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⠸⣿⣿⣿⣿⣿⣿⣿⣿⠟⠁⠀⣠⣾⣿⣿⣿⣦⡀⠈⠻⣿⣿⣿⣿⣿⣿⣿⣿⡏
⠀⢻⣿⣿⣿⣿⣿⠟⠁⠀⣠⣾⣿⣿⣿⣿⣿⣿⣿⣦⡀⠈⠻⣿⣿⣿⣿⣿⡿⠀
⠀⠀⢻⣿⣿⣟⣁⣀⣠⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⣀⣈⣻⣿⣿⡿⠁⠀
⠀⠀⠀⠙⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠋⠀⠀⠀
⠀⠀⠀⠀⠀⠙⠿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠋⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠉⠛⠿⠿⣿⣿⣿⣿⣿⣿⠿⠿⠛⠋⠁⠀⠀⠀⠀⠀⠀⠀`,

	"default": `
⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣠⣤⣶⣾⣿⣿⣿⣿⣿⣿⣿⣿⣷⣶⣤⣄⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⢀⣤⣾⣿⣿⠿⠟⠛⠉⠉⠁⠀⠀⠉⠉⠉⠛⠻⠿⣿⣿⣷⣤⡀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⢀⣴⣿⣿⠟⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⠻⣿⣿⣦⡀⠀⠀⠀⠀
⠀⠀⠀⣴⣿⣿⠟⠁⠀⠀⠀⠀⠀⣀⣤⣶⣶⣾⣿⣿⣶⣿⣿⣾⣶⡄⠀⠀⠈⠻⣿⣿⣦⠀⠀⠀
⠀⢀⣾⣿⡿⠁⠀⠀⠀⠀⠀⣠⣾⣿⣿⣿⣿⡿⠿⠿⠿⠿⠿⠿⠿⠃⠀⠀⠀⠀⠈⢿⣿⣷⡀⠀
⠀⣾⣿⡟⠀⠀⠀⠀⠀⢀⣼⣿⣿⣿⠟⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢻⣿⣷⠀
⢸⣿⣿⠁⠀⠀⠀⠀⠀⣼⣿⣿⡿⠁⠀⠀⢀⣠⣤⣤⣤⣤⣄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⣿⣿⡇
⣿⣿⡇⠀⠀⠀⠀⠀⠀⣿⣿⣿⡇⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⠇⠀⠀⠀⠀⠀⠀⠀⠀⠀⢸⣿⣿
⣿⣿⠇⠀⠀⠀⠀⠀⠀⣿⣿⣿⡇⠀⠀⠀⣿⣿⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢸⣿⣿
⣿⣿⡇⠀⠀⠀⠀⠀⠀⣿⣿⣿⡇⠀⠀⠀⣿⣿⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢸⣿⣿
⣿⣿⣇⠀⠀⠀⠀⠀⠀⣿⣿⣿⡇⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⡆⠀⠀⠀⠀⠀⠀⠀⠀⠀⣸⣿⣿
⠸⣿⣿⡀⠀⠀⠀⠀⠀⢻⣿⣿⣷⡀⠀⠀⠈⠙⠛⠛⠛⠛⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣿⣿⠇
⠀⢻⣿⣷⡀⠀⠀⠀⠀⠈⢿⣿⣿⣿⣦⣀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣾⣿⡟⠀
⠀⠀⢻⣿⣷⣄⠀⠀⠀⠀⠀⠙⢿⣿⣿⣿⣿⣿⣶⣶⣶⣶⣶⣶⣶⡄⠀⠀⠀⠀⣠⣾⣿⠟⠀⠀
⠀⠀⠀⠙⣿⣿⣦⡀⠀⠀⠀⠀⠀⠉⠛⠿⠿⢿⣿⣿⣿⣿⣿⣿⠿⠃⠀⠀⢀⣴⣿⡿⠋⠀⠀⠀
⠀⠀⠀⠀⠈⠻⢿⣿⣷⣤⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣤⣾⣿⡿⠟⠁⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠙⠻⣿⣿⣷⣶⣤⣤⣀⣀⣀⣀⣀⣀⣤⣤⣶⣾⣿⣿⠟⠋⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠛⠛⠿⢿⣿⣿⣿⣿⣿⣿⡿⠿⠟⠛⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀`,
}

func fetchCoinData(coinID string) (*CoinData, error) {
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&ids=%s&price_change_percentage=1h,24h,7d", coinID)

	// Create request with User-Agent header (good API citizenship)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %v", err)
	}
	req.Header.Set("User-Agent", fmt.Sprintf("cryptofetch/%s", version))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("Sorry, you reached the rate limit by CoinGecko. Please wait a minute before trying again.")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response: %v", err)
	}

	var coins []CoinData
	if err := json.Unmarshal(body, &coins); err != nil {
		return nil, fmt.Errorf("Failed to parse JSON: %v", err)
	}

	if len(coins) == 0 {
		return nil, fmt.Errorf("Coin '%s' not found", coinID)
	}

	return &coins[0], nil
}

func formatPrice(price float64) string {
	if price >= 1 {
		return fmt.Sprintf("$%.2f", price)
	} else if price >= 0.01 {
		return fmt.Sprintf("$%.4f", price)
	} else {
		return fmt.Sprintf("$%.8f", price)
	}
}

func formatLargeNumber(num float64) string {
	if num >= 1_000_000_000_000 {
		return fmt.Sprintf("$%.2fT", num/1_000_000_000_000)
	} else if num >= 1_000_000_000 {
		return fmt.Sprintf("$%.2fB", num/1_000_000_000)
	} else if num >= 1_000_000 {
		return fmt.Sprintf("$%.2fM", num/1_000_000)
	} else if num >= 1_000 {
		return fmt.Sprintf("$%.2fK", num/1_000)
	}
	return fmt.Sprintf("$%.2f", num)
}

func formatPercentage(pct float64) string {
	color := colorReset
	sign := ""

	if pct > 0 {
		color = colorGreen
		sign = "+"
	} else if pct < 0 {
		color = colorRed
	}

	return fmt.Sprintf("%s%s%.2f%%%s", color, sign, pct, colorReset)
}

func getASCIIArt(coinID string) string {
	art, ok := asciiArt[coinID]
	if !ok {
		art = asciiArt["default"]
	}
	return art
}

func getCoinColor(coinID string) string {
	color, ok := coinColors[coinID]
	if !ok {
		color = coinColors["default"]
	}
	return color
}

func resolveCoinID(input string) string {
	input = strings.ToLower(strings.TrimSpace(input))

	// Otherwise, assume it's already a coin ID
	return input
}

func displayCoinInfo(coin *CoinData) {
	art := getASCIIArt(coin.ID)
	artLines := strings.Split(strings.TrimSpace(art), "\n")
	artColor := getCoinColor(coin.ID)

	// Prepare info lines
	infoLines := []string{
		fmt.Sprintf("%s%s%s (%s)", colorBold, coin.Name, colorReset, strings.ToUpper(coin.Symbol)),
		fmt.Sprintf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s", artColor, colorReset),
		"",
		fmt.Sprintf("%sPrice%s         %s", colorYellow, colorReset, formatPrice(coin.CurrentPrice)),
		fmt.Sprintf("%sMarket Cap%s    %s", colorYellow, colorReset, formatLargeNumber(coin.MarketCap)),
		fmt.Sprintf("%sMarket Rank%s   #%d", colorYellow, colorReset, coin.MarketCapRank),
		fmt.Sprintf("%s24h Volume%s    %s", colorYellow, colorReset, formatLargeNumber(coin.TotalVolume)),
		"",
		fmt.Sprintf("%s24h High%s     %s", colorYellow, colorReset, formatPrice(coin.High24h)),
		fmt.Sprintf("%s24h Low%s      %s", colorYellow, colorReset, formatPrice(coin.Low24h)),
		"",
		fmt.Sprintf("%sPrice Changes:%s", colorBold, colorReset),
		fmt.Sprintf("  1h          %s", formatPercentage(coin.PriceChange1h)),
		fmt.Sprintf("  24h         %s", formatPercentage(coin.PriceChange24h)),
		fmt.Sprintf("  7d          %s", formatPercentage(coin.PriceChange7d)),
	}

	// Display side by side
	maxLines := len(artLines)
	if len(infoLines) > maxLines {
		maxLines = len(infoLines)
	}

	// Add padding to art lines if needed
	for len(artLines) < maxLines {
		artLines = append(artLines, "")
	}

	// Add padding to info lines if needed
	for len(infoLines) < maxLines {
		infoLines = append(infoLines, "")
	}

	// Display side by side
	for i := 0; i < maxLines; i++ {
		artLine := ""
		if i < len(artLines) {
			artLine = artLines[i]
		}

		infoLine := ""
		if i < len(infoLines) {
			infoLine = infoLines[i]
		}

		fmt.Printf("%s%-45s%s  %s\n", artColor, artLine, colorReset, infoLine)
	}

	fmt.Println()
}

func printHelp() {
	help := `
cryptofetch - Display cryptocurrency information in neofetch style

USAGE:
    cryptofetch [FLAGS]

FLAGS:
    --currency, --cur <name>    Specify cryptocurrency to display (default: bitcoin)
                                Can only use coin name that is used on CoinGecko (bitcoin, ethereum, bitcoin-cash, etc.)
    --version                   Show version information
    --help                      Show this help message

EXAMPLES:
    cryptofetch                      # Default: Show Bitcoin info
    cryptofetch --cur ethereum       # Show Ethereum info
    cryptofetch --cur solana         # Show Solana info

SUPPORTED CRYPTOCURRENCIES:
    All cryptocurrencies that can be found on CoinGecko (https://www.coingecko.com/en)

MORE INFO:
	Visit yaman.pro for more projects and info.

GITHUB: 
	github.com/julianyaman/cryptofetch
`
	fmt.Println(help)
}

func printVersion() {
	fmt.Printf("cryptofetch version %s\n", version)
}

func main() {
	var currency string
	var showHelp bool
	var showVersion bool

	flag.StringVar(&currency, "currency", "bitcoin", "Cryptocurrency to display")
	flag.StringVar(&currency, "cur", "bitcoin", "Cryptocurrency to display (shorthand)")
	flag.BoolVar(&showHelp, "help", false, "Show help message")
	flag.BoolVar(&showVersion, "version", false, "Show version information")

	flag.Parse()

	if showHelp {
		printHelp()
		return
	}

	if showVersion {
		printVersion()
		return
	}

	// Check if both flags were set (in which case --currency takes precedence)
	currencySet := false
	curSet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "currency" {
			currencySet = true
		}
		if f.Name == "cur" {
			curSet = true
		}
	})

	// Use --currency if set, otherwise use --cur, otherwise default to bitcoin
	if !currencySet && curSet {
		// --cur was explicitly set, currency variable already has the value
	} else if !currencySet && !curSet {
		currency = "bitcoin"
	}

	currency = resolveCoinID(currency)

	fmt.Printf("\n%sFetching %s data...%s\n\n", colorCyan, currency, colorReset)

	coin, err := fetchCoinData(currency)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	displayCoinInfo(coin)
}
