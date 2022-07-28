// Floki.go project main.go
package main

import (
	//"encoding/json"
	"bufio"
	//"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/dariubs/percent"
	"github.com/knetic/govaluate"

	//"reflect"

	gecko "github.com/superoo7/go-gecko/v3"
)

var (
	startTime     time.Time
	basePrice     float64
	basePriceStr  string
	token         string  = "YOUR_TOKEN_HERE"
	channelID     string  = "YOUR_CHANNEL_ID_HERE
	debugMaxPrice float64 = 1000000.000
	debugMinPrice float64 = 0
)

func init() {
	startTime = time.Now()
	basePrice, basePriceStr = GetCoin("shiba-inu", "usd")
	//flag.StringVar(&token, "t", "", "Bot Token")
	//flag.StringVar(&channelID, "c", "", "Channel ID")
	//flag.Parse()
}

func main() {
	go GetPriceOnLoop()
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	fmt.Println(ReplCurrentTime("[B] Main init complete"))
	dg.AddHandler(messageCreate)
	dg.Identify.Intents = discordgo.IntentsGuildMessages
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection1,", err)
		return
	}

	commands := map[string]interface{}{
		"help":          ReplHelp,
		"clear":         ReplCls,
		"uptime":        ReplReturnUptime,
		"current_price": ReplCurrentPrice,
		"base_price":    ReplReturnBase,
		"debug_0":       Repl0Value,
		"debug_max":     ReplMaxValue,
	}
	reader := bufio.NewReader(os.Stdin)
	//printRepl()
	text := get(reader)
	for ; shouldContinue(text); text = get(reader) {
		if value, exists := commands[text]; exists {
			value.(func())()
		} else {
			printInvalidCmd(text)
		}
		//printRepl()
	}
	fmt.Println("Bye!")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "$shib" {
		_, str := GetCoin("shiba-inu", "usd")
		message := "@everyone Current price of Shiba-Inu is: $" + str
		fmt.Println(ReplCurrentTime("[I] User requested Shiba-Inu price"))
		s.ChannelMessageSend(channelID, message)
	}

	if m.Content == "$uptime" {
		fmt.Println(ReplCurrentTime("[I] User requested uptime"))
		s.ChannelMessageSend(channelID, "This session of Floki has been running for "+time.Since(startTime).Round(time.Second).String())
	}
}

func GetPriceOnLoop() {
	ticker := time.NewTicker(time.Minute)
	quit := make(chan struct{})
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session in loop,", err)
		return
	}
	fmt.Println(ReplCurrentTime("[B] Time loop init complete"))
	dg.Identify.Intents = discordgo.IntentsGuildMessages
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	for {
		select {
		case <-ticker.C:
			val, str := GetCoin("shiba-inu", "usd")
			if str == "err" {
				fmt.Println(ReplCurrentTime("[!] API error occurred, continuing."))
				continue
			}

			if val >= (basePrice * 1.05) {
				basePrice = val
				basePriceStr = fmt.Sprintf("%.8f", basePrice)
				updateMessage := "Price increase! Current price: $" + str
				dg.ChannelMessageSend(channelID, "@everyone "+updateMessage)
				fmt.Println(ReplCurrentTime("[+] " + updateMessage))
			} else if val <= (basePrice * .95) {
				basePrice = val
				basePriceStr = fmt.Sprintf("%.8f", basePrice)
				updateMessage := "Price Decreased. Current price: $" + str
				dg.ChannelMessageSend(channelID, "@everyone "+updateMessage)
				fmt.Println(ReplCurrentTime("[-] " + updateMessage))
			} else {
				fmt.Println(ReplCurrentTime("[~] Current Price is: " + str))
			}
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func GetCoin(coin string, currency string) (float64, string) {
	// Will be used later to support additional cryptos
	// Will likely re-write when it comes time
	coinFixed := []string{coin}
	currencyFixed := []string{currency}
	cg := gecko.NewClient(nil)
	price, err := cg.SimplePrice(coinFixed, currencyFixed)
	if err != nil {
		log.Fatal(err)
		return 0, "err"
	}

	priceValue := float64((*price)[coinFixed[0]][currencyFixed[0]])
	priceString := fmt.Sprintf("%.8f", priceValue)

	return priceValue, priceString
}

func printRepl() {
	fmt.Print("> ")
}

func recoverExp(text string) {
	if r := recover(); r != nil {
		fmt.Println("> unknow command ", text)
	}
}

func printInvalidCmd(text string) {
	// We might have a panic here we so need DEFER + RECOVER
	defer recoverExp(text)
	// \n Will be ignored
	t := strings.TrimSuffix(text, "\n")
	if t != "" {
		expression, errExp := govaluate.NewEvaluableExpression(text)
		result, errEval := expression.Evaluate(nil)
		// Before we need to know if is not a Math expr
		if errExp == nil && errEval == nil {
			fmt.Println(result)
		} else {
			fmt.Println("> unknow command " + t)
		}
	}
}

func get(r *bufio.Reader) string {
	t, _ := r.ReadString('\n')
	return strings.TrimSpace(t)
}

func shouldContinue(text string) bool {
	if strings.EqualFold("exit", text) {
		return false
	}
	return true
}

func ReplHelp() {
	fmt.Println("These are the avaliable commands: ")
	fmt.Println("help   - Shows you this page")
	fmt.Println("cls    - Clear the Terminal Screen ")
	fmt.Println("exit   - Exits the Go REPL ")
	fmt.Println("1 + 2  - Its possible todo Math expressions: true == true, 4 * 6 / 2, 2 > 1 ")
	fmt.Println("uptime - Prints the uptime of this program ")
	fmt.Println("current_price - Shows the current price of Shiba-Inu")
	fmt.Println("base_price - Shows the base price of Shiba-Inu (used for calculating % change)")
}

func ReplCls() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
func ReplCurrentPrice() {
	val, price := GetCoin("shiba-inu", "usd")
	fmt.Println("^- ", price)
	fmt.Println("^- ", GetPercentage(val, basePrice), "% change from the base price of: ", basePriceStr)
}
func ReplReturnBase() {
	fmt.Println("^- ", basePriceStr)
}
func ReplReturnUptime() {
	//fmt.Println("^- ", time.Since(startTime))
	fmt.Println("^- ", time.Since(startTime).Round(time.Second))
}

func ReplCurrentTime(message string) string {
	time := time.Now().Round(time.Second)
	formattedTime := fmt.Sprintf("%02d-%02d-%d %02d:%02d:%02d",
		time.Month(), time.Day(), time.Year(),
		time.Hour(), time.Minute(), time.Second())
	return "[" + formattedTime + "]" + message
}

func GetPercentage(x, y float64) string {
	//percentageRounded := fmt.Sprintf("%.4f", (math.Abs(((x-y)/y)*100))) + "%"
	percentageRounded := fmt.Sprintf("%f", percent.ChangeFloat(y, x))
	return percentageRounded
}

func Repl0Value() {
	basePrice = debugMinPrice
	basePriceStr = fmt.Sprintf("%.8f", basePrice)
	fmt.Println(ReplCurrentTime("[D] Set base price to 0"))
}

func ReplMaxValue() {
	basePrice = debugMaxPrice
	basePriceStr = fmt.Sprintf("%.8f", basePrice)
	fmt.Println(ReplCurrentTime("[D] Set base price to 1,000,000"))
}
