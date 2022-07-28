package main

import (
	"fmt"
	"errors"
	"strings"
	"strconv"
)

/**
jday is the numeric value of a date, where 0 is January 1, 1901.  This is only valid through
December 31, 2099. We exclude the years 1900 and 2100 because they are oddball years - (they are divisible
by 4 but they are not leap years).

qday is the day number in a 4 year cycle from 0..1460.

jday 1461 is January 1, 1905
*/

/**
* IsoDate is a string in the format YYYY-MM-DD
* See https://en.wikipedia.org/wiki/ISO_8601.
* We allow dates that aren't strictly correct, like 2022-02-31 would be legal
* This is just a wrapper around the date string
*/
type IsoDate struct {
	iso string
}

func (d *IsoDate) toString() string {
	return d.iso
}

func newIsoDate(yyyy int, mm int, dd int) (*IsoDate,error) {
	if yyyy < 1901 || yyyy > 2099 {
		return nil, errors.New("invalid year: " + strconv.Itoa(yyyy))
	} else if  mm < 1 || mm > 12 {
		return nil, errors.New("invalid month: " + strconv.Itoa(mm))
	} else if  dd < 1 || dd > 31 {
		return nil, errors.New("invalid day: " + strconv.Itoa(dd))
	} else {
		s := fmt.Sprintf("%d-%02d-%02d",yyyy,mm,dd)
		return &IsoDate{iso: s}, nil
	}
}

/**
* Input a string, validate it and create a new ISO date
*/
func newIsoDateFromString(isod string) (*IsoDate,error) {
	if isod == "" {
		return nil, errors.New("invalid date string: '' ")
	}
	res := strings.Split(isod, "-")
	if len(res)!=3 {
		return nil, errors.New("invalid date string: "+isod)
	}
	y, err := strconv.Atoi(res[0])
	if (err!=nil) {return nil,err}
	m, err := strconv.Atoi(res[1])
	if (err!=nil) {return nil,err}
	d, err := strconv.Atoi(res[2])
	if (err!=nil) {return nil,err}
	return newIsoDate(y, m, d)
}

//we don't need to error check this because we have already done that
func (d *IsoDate) parts() (int,int,int) {
	res := strings.Split(d.iso, "-")
	y, _ := strconv.Atoi(res[0])
	m, _ := strconv.Atoi(res[1])
	dd, _ := strconv.Atoi(res[2])
	return y,m,dd
}

func (d *IsoDate) toJday() int {
	y,m,dd := d.parts()
	dpy := daysInPriorYears(y)
	dpm := daysInPriorMonths(y,m)
	return dpy + dpm + dd - 1
}

//like it says, return the number of days in prior years
//where the year 1901 would return 0
// y must be in the range 1901..2099, but we have already validated it
func daysInPriorYears(y int) int {
	//cycles means number of complete 1461-day cycles
	cycles := ((y - 1901) / 4) 
	//cur will be a value from 0..3, where 3 means divisible by 4 and is a leap year
	//1901 will have a value of 0
	cur := (y - 1901) % 4
	if (cur==3) {
		return (cycles * 1461) + 1095
	} else {
		return (cycles * 1461) + (cur * 365)
	}
}

//this holds the days year to date through the beginning of the month
//to use this, subtract 1 from the current month and get the value.
//for example, Feb is the 2nd month, so 1 should be 31
var mdays = [12]int{0,31,59,90,120,151,181,212,243,273,304,334}

//return the days in prior months.  We need the years to determine if it is a leap year
func daysInPriorMonths(y int,m int) int {
	cur := (y - 1901) % 4
	if (cur==3 && m>2) {
		return mdays[m-1]+1
	} else {
		return mdays[m-1]
	}
}

//----------------------------------------------
//the following functions operate on the aforementioned jday

// Saturday is both 0 and 7
//this goes up to 9 to make it easier with a function below
var Weekdays = [10]string{"Saturday","Sunday","Monday","Tuesday","Wednesday",
	"Thursday","Friday","Saturday","Sunday","Monday"}

//return DayOfWeek where i is 0..7
func DayOfWeek(jday int) string {
	dow := (jday % 7) + 3
	return Weekdays[dow]
}

//given the jday, return the year
func jyear(jday int) int {
	q := jday / 1461
	r := jday % 1461
	t := r / 365
	if (r==1460) {
		//handle the oddball case of 12/31/04
		return 1901 + (q * 4) + 3
	} else {
		return 1901 + (q * 4) + t
	}
}

//given the jday, return the month as a value from 1..12
//this repeats some code from above, this could be combined at some point
func jmonthday(jday int) (int,int) {
	r := jday % 1461
	u := r % 365
	//handle the oddball case of 12/31/04
	if r==1460 {
		return 12,31
	} else {
		//if it is a leap year
		//and it is after 2/28
		if r>1153 {	
			if (u<60) {
				return 2, (u-30)
			} else if u < 91 {
				return 3, (u-59)
			} else if u < 121 {
				return 4, (u-90)
			} else if u < 152 {
				return 5, (u-120)
			} else if u < 182 {
				return 6, (u-151) 
			} else if u < 213 {
				return 7, (u-181) 
			} else if u < 244 {
				return 8, (u-212) 
			} else if u < 274 {
				return 9, (u-243) 
			} else if u < 305 {
				return 10, (u-273) 
			} else if u < 335 {
				return 11, (u-304)
			} else {
				return 12,(u-334)
			}	
		}
	}

	//implicit else
	if u < 31 {
		return 1, (u+1)
	} else if u < 59 {
		return 2, (u-31+1)
	} else if u < 90 {
		return 3, (u-59+1)
	} else if u < 120 {
		return 4, (u-90+1)
	} else if u < 151 {
		return 5, (u-120+1)
	} else if u < 181 {
		return 6, (u-151+1) 
	} else if u < 212 {
		return 7, (u-181+1) 
	} else if u < 243 {
		return 8, (u-212+1) 
	} else if u < 273 {
		return 9, (u-243+1) 
	} else if u < 304 {
		return 10, (u-273+1) 
	} else if u < 334 {
		return 11, (u-304+1)
	} else {
		return 12,(u-334+1)
	}
}

//----------------------------------------------
func main() {
	var isod string
	fmt.Print("Enter a date in the format YYYY-MM-DD: ")
  	fmt.Scanln(&isod)
	iso, err := newIsoDateFromString(isod)
	if (err!=nil) {
		fmt.Println(err)
		return
	} 
	fmt.Println(iso.toString())
	j := iso.toJday()
	fmt.Println(j)
	//------------
	jy := jyear(j)
	jm,jd := jmonthday(j)
	//fmt.Println(jy,jm,jd)
	iso2, err := newIsoDate(jy,jm,jd)
	fmt.Print(DayOfWeek(j))
	fmt.Println(", "+iso2.toString())
}
