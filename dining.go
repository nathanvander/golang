package main

//This is the Dining Philosophers problem from http://www.rosettacode.org/wiki/Dining_philosophers
//it is used to test deadlocks

import (
    "fmt"
    "sync"
    //"math/rand"
    "time"    
)

//time constants
const (
	//duration is in nanoseconds
	millis time.Duration = 1000000
)

//Philosopher state
const (
	Ready = 0
	Hungry = 1		//trying to get a fork
	Eating = 2
	Pondering = 3
	Done = 4
)

type Philosopher struct {
	id int		//number from 0..6
	name string
	state int	//one of the states above
	course int
	leftfork int
	rightfork int
}

//list of philosophers
var phers [7]*Philosopher

//constructor.  return new Philosopher object
func NewPhilosopher(num int,nam string,left int,right int) *Philosopher {
	ph := &Philosopher{
		id: num,
		name: nam,
		leftfork: left,
		rightfork: right,
	}
	//add it to the list of philosophers
	phers[num] = ph
	fmt.Println(ph.name + " has joined the table")
	return ph
}

//methods
func (ph *Philosopher) ID() int {return ph.id }
func (ph *Philosopher) Name() string {return ph.name}
func (ph *Philosopher) State() int { return ph.state}
func (ph *Philosopher) Course() int {return ph.course} 
func (ph *Philosopher) LeftFork() int {return ph.leftfork} 
func (ph *Philosopher) RightFork() int {return ph.rightfork}

//this is the run method
func (ph *Philosopher) Dine() {
	time.Sleep(100 * millis)
	for ; ph.course < 4; {
		ph.state = Hungry
		fmt.Println(ph.Name() + " is hungry")	
		//This is very simple. Eat, ponder, eat, ponder, eat, done
		left := GetFork(ph.LeftFork())
		right := GetFork(ph.RightFork())
		left.Pickup(ph);
		right.Pickup(ph);
		fmt.Println(ph.name + " is eating")
		ph.state = Eating
		time.Sleep(100 * millis)
		left.Drop(ph)
		right.Drop(ph)
		ph.state = Pondering
		fmt.Println(ph.name + " is pondering")
		time.Sleep(100 * millis)
		ph.course++
	}
	fmt.Println(ph.name + " is full")
	ph.state = Done
}

//===========================

type Fork struct {
	id int
	//mu is a pointer to the mutex
	mu *sync.Mutex
}

//list of forks
var forks [7]*Fork

func GetFork(num int) *Fork { return forks[num];} 

//constructor
func NewFork(num int) *Fork {
	f := &Fork{
		id: num,
		//create a new mutex
		mu: &sync.Mutex{},
	}
	//add it to the list of forks
	forks[num] = f
	fmt.Print("fork ")
	fmt.Print(num)
	fmt.Println(" is placed on the table")
	return f
}
func (f *Fork) ID() int {return f.id }
func (f *Fork) Pickup(ph *Philosopher) {
	f.mu.Lock()
	fmt.Print(ph.Name()+" has picked up fork ")
	fmt.Println(f.ID())
}
func (f *Fork) Drop(ph *Philosopher) {
	//fmt.Print(ph.Name()+" is attempting to drop fork ")
	//fmt.Println(f.ID())
	f.mu.Unlock()	
	fmt.Print(ph.Name()+" has dropped fork ")
	fmt.Println(f.ID())
}


//===========================
//test code
func main() {
	fmt.Println("Dining Philosophers")
	fmt.Println("There are 5 philosophers and 5 forks.")
	fmt.Println("A philosophers needs 2 forks to eat.")
	fmt.Println("Can they cooperate by sharing forks?")
	//create forks
	for i:=0;i<5;i++ {
		NewFork(i)
	}
	
	//create philosophers
	ph0 := NewPhilosopher(0,"Aristotle",0,1)
	ph1 := NewPhilosopher(1,"Kant",1,2)	
	ph2 := NewPhilosopher(2,"Marx",2,3)	
	ph3 := NewPhilosopher(3,"Plato",3,4)
	ph4 := NewPhilosopher(4,"Russell",4,0)	
	//ph5 := NewPhilosopher(5,"Spinoza",5,6)
	//ph6 := NewPhilosopher(6,"Nietzsche",6,0)	
	go ph0.Dine()
	go ph1.Dine()
	go ph2.Dine()
	go ph3.Dine()
	go ph4.Dine()
	//go ph5.Dine()
	//go ph6.Dine()
	time.Sleep(2*time.Second)
}