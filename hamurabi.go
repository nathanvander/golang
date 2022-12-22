package main

//The Java source is from: https://www.cis.upenn.edu/+matuszek/cit590-2009/Examples/Hammurabi.java
//another version is at https://github.com/dryack/go-hamurabi/

import (
	"fmt"
	//"math"
	"math/rand"
	"time"
	"strconv"
)

//game variables
var	random *rand.Rand
var	year int
var	population int
var	grain int
var	acres int
var	landValue int
var	starved int
var	percentStarved int
var	plagueVictims int
var	immigrants int
var	grainHarvested int
var	harvestPerAcre int
var	amountEatenByRats int
var	grainFedToPeople int
var	acresPlanted int

const OGH string = "O Great Hammurabi!"

	//This is the random number function
	//int get_random(int range) {
	//	return uniform(0, range , rand);
	//}

func init_randomizer() {
 	s1 := rand.NewSource(time.Now().UnixNano())
    random = rand.New(s1)
}

func get_random(n int) int {
	v := random.Intn(n)
	return v;
}

func getNumber(message string) int {
	var input string
	for true {	//while true
    	fmt.Println(message)
    	fmt.Scanln(&input)
        n, err := strconv.Atoi(input)
        
        if err != nil {
        	fmt.Println("Please enter a number")
        } else {
        	return n
        }
    }
    return 0
}

func jest(message string) {
	fmt.Println(OGH + ", surely you jest!");
    fmt.Println(message);
}


//starts with an homage to the basic source code
func printIntroductoryParagraph() {
	fmt.Println(`
HAMURABI 
CREATIVE COMPUTING  MORRISTOWN, NEW JERSEY.

Congratulations! You are the newest ruler of ancient Samaria,
elected for a ten year term of office. Your duties are to
dispense food, direct farming, and buy and sell land as
needed to support your people. Watch out for rat infestations
and the plague! Grain is the general currency, measured in
bushels.

The following will help you in your decisions:                
   * Each person needs at least 20 bushels of grain per year to survive
   * Each person can farm at most 10 acres of land
   * It takes 2 bushels of grain to farm an acre of land
   * The market price for land fluctuates yearly
     
Rule wisely and you will be showered with appreciation at the
end of your term. Rule poorly and you will be kicked out of office!
	`)
}

//--------------------------------
func playGame() {
		var stillInOffice bool = true;

        initializeVariables();
        printSummary();
        for year <= 10 && stillInOffice {
            buyLand();
            sellLand();
            feedPeople();
            plantGrain();

            checkForPlague();
            countStarvedPeople();
            if (percentStarved >= 45) {
                stillInOffice = false;
            }
            countImmigrants();
            takeInHarvest();
            checkForRats();
            updateLandValue();
            printSummary();
            year = year + 1;
        }
        printFinalScore();
    }

    /**
     * Initialize all instance variables for start of game.
     */

    func initializeVariables() {
        year = 1;
        population = 100;
        grain = 2800;
        acres = 1000;
        landValue = 19;
        starved = 0;
        plagueVictims = 0;
        immigrants = 5;
        grainHarvested = 3000;
        harvestPerAcre = 3;
        amountEatenByRats = 200;
    }

    /**
     * Prints the year-end summary.
     */
	func printSummary() {
        fmt.Println("___________________________________________________________________");
        fmt.Println(OGH);
        fmt.Println("You are in year " + strconv.Itoa(year) + " of your ten year rule.");
        if (plagueVictims > 0) {
            fmt.Println("A horrible plague killed " + strconv.Itoa(plagueVictims) + " people.");
        }
        fmt.Println("In the previous year " + strconv.Itoa(starved) + " people starved to death,");
        fmt.Println("and " + strconv.Itoa(immigrants) + " people entered the kingdom.");
        fmt.Println("The population is now " + strconv.Itoa(population) + ".");
        fmt.Println("We harvested " + strconv.Itoa(grainHarvested) + " bushels at " + strconv.Itoa(harvestPerAcre) + " bushels per acre.");
        if (amountEatenByRats > 0) {
            fmt.Println("*** Rats destroyed " + strconv.Itoa(amountEatenByRats) + " bushels, leaving " + strconv.Itoa(grain) + " bushels in storage.");
        } else {
            fmt.Println("We have " + strconv.Itoa(grain) + " bushels of grain in storage.");
        }
        fmt.Println("The city owns " + strconv.Itoa(acres) + " acres of land.");
        fmt.Println("Land is currently worth " + strconv.Itoa(landValue) + " bushels per acre.");
        fmt.Println();
    }

    /**
     * Allows the user to buy land.
     */
    func buyLand() {
        var acresToBuy int;
        var question string = "How many acres of land will you buy? ";
		var cost int;

        acresToBuy = getNumber(question);
        cost = landValue * acresToBuy;
        for (cost > grain) {
            jest("We have but " + strconv.Itoa(grain) + " bushels of grain, not " + strconv.Itoa(cost) + "!");
            acresToBuy = getNumber(question);
            cost = landValue * acresToBuy;
        }
        grain = grain - cost;
        acres = acres + acresToBuy;
        fmt.Println(OGH + ", you now have " + strconv.Itoa(acres) + " acres of land");
        fmt.Println("and " + strconv.Itoa(grain) + " bushels of grain.");
    }



    /**
     * Allows the user to sell land.
     */
    func sellLand() {
        var question string = "How many acres of land will you sell? ";
        var acresToSell int = getNumber(question);

        for (acresToSell > acres) {
            jest("We have but " + strconv.Itoa(acres) + " acres!");
            acresToSell = getNumber(question);
        }
        grain = grain + landValue * acresToSell;
        acres = acres - acresToSell;
        fmt.Println(OGH + ", you now have " + strconv.Itoa(acres) + " acres of land");
        fmt.Println("and " + strconv.Itoa(grain) + " bushels of grain.");
    }

    /**
     * Allows the user to decide how much grain to use to feed people.
     */
    func feedPeople() {
        var question string = "How much grain will you feed to the people? ";
        grainFedToPeople = getNumber(question);

        for (grainFedToPeople > grain) {
            jest("We have but " + strconv.Itoa(grain) + " bushels!");
            grainFedToPeople = getNumber(question);
        }
        grain = grain - grainFedToPeople;
        fmt.Println(OGH + ", " + strconv.Itoa(grain) + " bushels of grain remain.");
    }

    /**
     * Allows the user to choose how much grain to plant.
     */
    func plantGrain() {
        var question string = "How many bushels will you plant? ";
        var amountToPlant int = 0;
        var haveGoodAnswer bool = false;

        for (!haveGoodAnswer) {
            amountToPlant = getNumber(question);
            if (amountToPlant > grain) {
                jest("We have but " + strconv.Itoa(grain) + " bushels left!");
            } else if (amountToPlant > 2 * acres) {
                jest("We have but " + strconv.Itoa(acres) + " acres available for planting!");
            } else if (amountToPlant > 20 * population) {
                jest("We have but " + strconv.Itoa(population) + " people to do the planting!");
            } else {
                haveGoodAnswer = true;
            }
        }
        acresPlanted = amountToPlant / 2;
        grain = grain - amountToPlant;
        fmt.Println(OGH + ", we now have " + strconv.Itoa(grain) + " bushels of grain in storage.");
    }

    /**
     * Checks for plague, and counts the victims.
     */
    func checkForPlague() {
        var chance int = get_random(100);
        if (chance < 15) {
        	fmt.Println("*** A horrible plague kills half your people! ***");
            plagueVictims = population / 2;
            population = population - plagueVictims;
        } else {
            plagueVictims = 0;
        }
    }

    /**
     * Counts how many people starved, and removes them from the population.
     */
    func countStarvedPeople() {
        var peopleFed int = grainFedToPeople / 20;
        if (peopleFed >= population) {
            starved = 0;
            percentStarved = 0;
            fmt.Println("Your people are well fed and happy.");
        } else {
            starved = population - peopleFed;
            fmt.Println( strconv.Itoa(starved) + " people starved to death.");
            percentStarved = (100 * starved) / population;
            population = population - starved;
        }
    }

    /**
     * Counts how many people immigrated.
     */
    func countImmigrants() {
        if (starved > 0) {
            immigrants = 0;
        } else {
            immigrants = (20 * acres + grain) / (100 * population) + 1;
            population += immigrants;
        }
    }

    /**
     * Determines the harvest, and collects the new grain.
     */
    func takeInHarvest() {
        harvestPerAcre = get_random(5) + 1;
        grainHarvested = harvestPerAcre * acresPlanted;
        grain = grain + grainHarvested;
    }

    /**
     * Checks if rats get into the grain, and determines how much they eat.
     */
    func checkForRats() {
        if (get_random(100) < 40) {
            var percentEatenByRats int = 10 + get_random(21);
            fmt.Println("*** Rats eat " + strconv.Itoa(percentEatenByRats) + " percent of your grain! ***");
            amountEatenByRats = (percentEatenByRats * grain) / 100;
            grain = grain - amountEatenByRats;
        } else {
            amountEatenByRats = 0;
        }
    }

    /**
     * Randomly sets the new price of land.
     */
    func updateLandValue() {
        landValue = 17 + get_random(7);
    }

    /**
     * Prints an evaluation at the end of a game.
     */
    func printFinalScore() {
        if (starved >= (45 * population) / 100) {
        	fmt.Println("O Once-Great Hammurabi");
			fmt.Println(strconv.Itoa(starved) + " of your people starved during the last year of your");
			fmt.Println("incompetent reign! The few who remain have stormed the palace");
            fmt.Println("and bodily evicted you!");
            fmt.Println("\nYour final rating: TERRIBLE.");
            return;
        }
        var plantableAcres int = acres;
        if (20 * population < plantableAcres) {
            plantableAcres = 20 * population;
        }

        if (plantableAcres < 600) {
            fmt.Println("Congratulations, " + OGH);
            fmt.Println(" You have ruled wisely but not");
            fmt.Println("well; you have led your people through ten difficult years, but");
            fmt.Println("your kingdom has shrunk to a mere " + strconv.Itoa(acres));
            fmt.Println(" acres.\n" + "\nYour final rating: ADEQUATE.");
        } else if (plantableAcres < 800) {
            fmt.Println("Congratulations, \" + OGH + \" You  have ruled wisely, and");
            fmt.Println("shown the ancient world that a stable economy is possible.");
            fmt.Println("\nYour final rating: GOOD.");
        } else {
            fmt.Println("Congratulations, " + OGH + " You  have ruled wisely and well, and");
            fmt.Println("expanded your holdings while keeping your people happy.");
            fmt.Println("Altogether, a most impressive job!");
            fmt.Println("\nYour final rating: SUPERB.");
        }
    }


//=====================	
func main() {
	init_randomizer()
	printIntroductoryParagraph()
	playGame()
}

