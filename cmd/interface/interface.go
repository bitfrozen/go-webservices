package main

import "fmt"

// Wallet
type Wallet struct {
	Cash int
}

func (w *Wallet) Pay(amount int) error {
	if w.Cash < amount {
		return fmt.Errorf("not enough cash")
	}
	w.Cash -= amount
	return nil
}

// Card
type Card struct {
	Balance    int
	ValidUntil string
	Cardholder string
	CVV        string
	Number     string
}

func (c *Card) Pay(amount int) error {
	if c.Balance < amount {
		return fmt.Errorf("not enough money on card")
	}
	c.Balance -= amount
	return nil
}

// ApplePay
type ApplePay struct {
	Money   int
	AppleID string
}

func (a *ApplePay) Pay(amount int) error {
	if a.Money < amount {
		return fmt.Errorf("not enough money on account")
	}
	a.Money -= amount
	return nil
}

type Payer interface {
	Pay(int) error
}

func Buy(p Payer) {
	switch p.(type) {
	case *Wallet:
		fmt.Println("Paying cash?")
	case *Card:
		plasticCard, ok := p.(*Card)
		if !ok {
			fmt.Println("Can't cast to type *Card")
		}
		fmt.Println("Insert card,", plasticCard.Cardholder)
	default:
		fmt.Println("Something new")
	}
	err := p.Pay(10)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Thank you for using %T\n\n", p)
}

func main() {
	myWallet := &Wallet{Cash: 100}
	Buy(myWallet)

	var myCard Payer
	myCard = &Card{Balance: 100, Cardholder: "bitfrozen"}
	Buy(myCard)

	myApple := &ApplePay{Money: 9}
	Buy(myApple)
}
