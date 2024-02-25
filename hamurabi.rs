//Hamurabi
/** This requires the following Cargo.tomlL
-------------
[package]
name = "hamurabi"
version = "0.1.0"
edition = "2021"

[dependencies]
rand = "0.8.5"

[[bin]]
name = "hamurabi"
path = "src/hamurabi.rs"
-------------
*/

use rand::Rng;

//static variables
static mut year: i32 = 0;
static mut population: i32 = 0;
static mut grain: i32 = 0;
static mut acres: i32 = 0;
static mut landValue: i32 = 0;
static mut starved: i32 = 0;
static mut percentStarved: i32 = 0;
static mut plagueVictims: i32 = 0;
static mut immigrants: i32 = 0;
static mut grainHarvested: i32 = 0;
static mut harvestPerAcre: i32 = 0;
static mut amountEatenByRats: i32 = 0;
static mut grainFedToPeople: i32 = 0;
static mut acresPlanted: i32 = 0;
static mut stillInOffice: bool = true;
static OGH:&str = "O Great Hammurabi!";

//--------------------------------

//utility functions
//get random number from 0 .. n-1
fn get_random(n: i32) -> i32 {
	let mut rng = rand::thread_rng();
	let r: f64 = rng.gen::<f64>();
	let f2: f64 = r * (n as f64);
	let a: i32 = f2 as i32;	
	return a;
}

//get number from user
fn get_number() -> i32 {
	let mut buff = String::new();
	let stdin = std::io::stdin(); 
	let _ = stdin.read_line(&mut buff);
	let buff = buff.trim();
	let n: i32 = buff.parse::<i32>().expect("Not a valid number");
	return n;
}

fn jest() {
	println!("{}, surely you jest!", OGH);
}

//-----------------------
//starts with an homage to the basic source code
fn print_intro() {
	let intro = r#"	
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

(Hint: You should feed 20 bushels of grain per person 
and plant 2 bushel per acre. Each person can farm 10 acres)
	"#;
	println!("{}", intro);
}

unsafe fn playGame() {
    initializeVariables();
    //printSummary();
    if stillInOffice {
        for _y in 1..10 {
        	println!("Year: {}",year);
        	printSummary();
            buyLand();
            sellLand();
            feedPeople();
            plantGrain();

            checkForPlague();
            countStarvedPeople();
            if percentStarved >= 45 {
                stillInOffice = false;
            }
            countImmigrants();
            takeInHarvest();
            checkForRats();
            updateLandValue();
            year = year + 1;
        }
	}
    printFinalScore();
}

    /**
     * Initialize all instance variables for start of game.
     */

    unsafe fn initializeVariables() {
        year = 1;
        population = 100;
        //initial grain will be from 2000..4500
        grain = get_random(2500) + 2000;
        acres = 1000;
        landValue = 19;
        starved = 0;
        plagueVictims = 0;
        immigrants = 5;
        grainHarvested = 3000;
        harvestPerAcre = 3;
        amountEatenByRats = 200;
        stillInOffice = true;
    }

    /**
     * Prints the year-end summary.
     */
	unsafe fn printSummary() {
        println!("___________________________________________________________________");
        println!("{}",OGH);
        println!("You are in year {} of your ten year rule.", year);
        if plagueVictims > 0 {
            println!("A horrible plague killed {} people.", plagueVictims);
        }
        println!("In the previous year {} people starved to death,",starved);
        println!("and {} people entered the kingdom.",immigrants);
        println!("The population is now {}.",population);
        println!("We harvested {} bushels at {} bushels per acre.",grainHarvested,harvestPerAcre);
        if amountEatenByRats > 0 {
            println!("*** Rats destroyed {} bushels, leaving {} bushels in storage.",amountEatenByRats,grain);
        } else {
            println!("We have {} bushels of grain in storage.", grain);
        }
        println!("The city owns {} acres of land.", acres);
        println!("Land is currently worth {} bushels per acre.", landValue);
        println!();
    }

    /**
     * Allows the user to buy land.
     */
    unsafe fn buyLand() {
        let mut acresToBuy: i32 = 0;
        let mut cost: i32 = 0;
        let q: String = "How many acres of land will you buy? ".to_string();
        
        println!("{}",q);
        acresToBuy = get_number();
        cost = landValue * acresToBuy;
        while cost > grain {
        	jest();
            println!("We have but {} bushels of grain, not {} !",grain,cost);
            println!("{}",q);
            acresToBuy = get_number();
            cost = landValue * acresToBuy;
        }
        grain = grain - cost;
        acres = acres + acresToBuy;
        println!("{}, you now have {} acres of land",OGH,acres);
        println!("and {} bushels of grain.",grain);
    }


    /**
     * Allows the user to sell land.
     */
    unsafe fn sellLand() {
    	let q: String = "How many acres of land will you sell? ".to_string();
    	println!("{}",q);
    	let mut acresToSell: i32 = get_number();

        while acresToSell > acres {
            jest();
            println!("We have but {} acres!",acres);
            println!("{}",q);
            acresToSell = get_number();
        }
        grain = grain + landValue * acresToSell;
        acres = acres - acresToSell;
        println!("{}, you now have {} acres of land",OGH,acres);
        println!("and {} bushels of grain.",grain);
    }

    /**
     * Allows the user to decide how much grain to use to feed people.
     */
    unsafe fn feedPeople() {
        let q: String = "How much grain will you feed to the people? ".to_string();
        println!("{}",q);
        grainFedToPeople = get_number();

		while grainFedToPeople > grain {
            jest();
            println!("We have but {} bushels!",grain);
            println!("{}",q);
        	grainFedToPeople = get_number();
        }
        grain = grain - grainFedToPeople;
        println!("{}, {} bushels of grain remain.",OGH,grain);
    }

    /**
     * Allows the user to choose how much grain to plant.
     */
    unsafe fn plantGrain() {
        let q: String = "How many bushels will you plant? ".to_string();
        let mut amountToPlant: i32 = 0;
        let mut haveGoodAnswer: bool = false;

        while !haveGoodAnswer {
        	println!("{}",q);
            amountToPlant = get_number();
            if amountToPlant > grain {
                jest();
                println!("We have but {} bushels left!",grain);
            } else if amountToPlant > 2 * acres {
                jest();
                println!("We have but {} acres available for planting!",acres);
            } else if (amountToPlant > 20 * population) {
                jest();
                println!("We have but {} people to do the planting!",population);
            } else {
                haveGoodAnswer = true;
            }
        }
        acresPlanted = amountToPlant / 2;
        grain = grain - amountToPlant;
        println!("{}, we now have {} bushels of grain in storage.",OGH,grain);
    }

    /**
     * Checks for plague, and counts the victims.
     */
    unsafe fn checkForPlague() {
        let chance: i32 = get_random(100);
        if (chance < 15) {
        	println!("*** A horrible plague kills half your people! ***");
            plagueVictims = (population / 2) as i32;
            population = population - plagueVictims;
        } else {
            plagueVictims = 0;
        }
    }

    /**
     * Counts how many people starved, and removes them from the population.
     */
    unsafe fn countStarvedPeople() {
        let peopleFed: i32 = (grainFedToPeople / 20) as i32;
        if (peopleFed >= population) {
            starved = 0;
            percentStarved = 0;
            println!("Your people are well fed and happy.");
        } else {
            starved = population - peopleFed;
            println!("{} people starved to death.",starved);
            percentStarved = ((100 * starved) / population) as i32;
            population = population - starved;
        }
    }

    /**
     * Counts how many people immigrated.
     */
   unsafe fn countImmigrants() {
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
    unsafe fn takeInHarvest() {
        harvestPerAcre = get_random(6) + 1;
        grainHarvested = harvestPerAcre * acresPlanted;
        grain = grain + grainHarvested;
    }

    /**
     * Checks if rats get into the grain, and determines how much they eat.
     */
    unsafe fn checkForRats() {
        if (get_random(100) < 40) {
            let percentEatenByRats: i32 = 10 + get_random(21);
            println!("*** Rats eat {} percent of your grain! ***",percentEatenByRats);
            amountEatenByRats = ((percentEatenByRats * grain) / 100) as i32;
            grain = grain - amountEatenByRats;
        } else {
            amountEatenByRats = 0;
        }
    }

    /**
     * Randomly sets the new price of land.
     */
    unsafe fn updateLandValue() {
        landValue = 17 + get_random(7);
    }

    /**
     * Prints an evaluation at the end of a game.
     */
    unsafe fn printFinalScore() {
        if (starved >= (45 * population) / 100) {
        	println!("O Once-Great Hammurabi");
			println!("{} of your people starved during the last year of your",starved);
			println!("incompetent reign! The few who remain have stormed the palace");
            println!("and bodily evicted you!");
            println!("\nYour final rating: TERRIBLE.");
            return;
        }
        let mut plantableAcres: i32 = acres;
        if (20 * population < plantableAcres) {
            plantableAcres = 20 * population;
        }

        if (plantableAcres < 600) {
            println!("Congratulations, {}!",OGH);
            println!("You have ruled wisely but not");
            println!("well; you have led your people through ten difficult years, but");
            println!("your kingdom has shrunk to a mere {}",acres);
            println!(" acres.");
            println!("Your final rating: ADEQUATE.");
        } else if (plantableAcres < 800) {
            println!("Congratulations, {}! You  have ruled wisely, and",OGH);
            println!("shown the ancient world that a stable economy is possible.");
            println!("\nYour final rating: GOOD.");
        } else {
            println!("Congratulations, {} You  have ruled wisely and well, and",OGH);
            println!("expanded your holdings while keeping your people happy.");
            println!("Altogether, a most impressive job!");
            println!("\nYour final rating: SUPERB.");
        }
    }

//===============================
fn main() {
	print_intro();
	unsafe {
		playGame();
	}
}
