# How to Add New ASCII Art for Cryptocurrencies

To add new ASCII art for cryptocurrencies in Cryptofetch, follow these steps:

1. **Find or Create ASCII Art**:

- We recommend to use emojicombos Dot Art Generator: https://emojicombos.com/dot-art-generator
- Make sure, that the ASCII art has a similar style to the existing ones in the [main.go](../main.go) codebase.
- The generated art should have these properties:
  - threshold=0.50
  - width=30
  - characters=464

2. **Open `main.go`**:

- Go to the `asciiArt` map in the `main.go` file.

3. **Add New Entry**:

- Add a new entry to the `asciiArt` map with the cryptocurrency's CoinGecko ID as the key and the ASCII art as the value.

4. **Pull Request**:

- After adding the new ASCII art, create a pull request to the main repository with your changes.