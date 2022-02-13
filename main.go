/*
	Author		: Nancy Yang / maxbearwiz@gmail.com
	Date			: 2022-02-13
	Description : Sample Photographer Scheduler
*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

type TimeSlot struct {
	Id     string    `json:"id,omitempty"`
	Starts time.Time `json:"starts"`
	Ends   time.Time `json:"ends"`
}

type Photographer struct {
	Id             string     `json:"id"`
	Name           string     `json:"name"`
	Availabilities []TimeSlot `json:"availabilities,omitempty"`
	Bookings       []TimeSlot `json:"bookings,omitempty"`
}

type Schedule struct {
	Photographer Photographer `json:"photographer"`
	TimeSlot     TimeSlot     `json:"timeSlot"`
}

type Photographers struct {
	Photographers []Photographer `json:"photographers"`
	Schedules     []Schedule
}

type Scheduler interface {
	readInput(fname string)
	printInput()
	addAvailableTimeSlot(photographer Photographer, starts time.Time, durationInMinutes int)
	availableTimeSlotsForBooking(durationInMinutes int) []Schedule
}

func (photographers *Photographers) readInput(fname string) (err error) {
	jsonFile, err := os.Open(fname)
	if err != nil {
		fmt.Println(err)
		return
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal(byteValue, photographers)
	return
}

func (photographers *Photographers) printInput() {
	fmt.Printf("Photographers:\n")
	for i := 0; i < len(photographers.Photographers); i++ {
		var photographer = photographers.Photographers[i]
		fmt.Printf("\tid  : %s\n", photographer.Id)
		fmt.Printf("\tname: %s\n", photographer.Name)

		fmt.Printf("\tavailabilities\n")
		for ii := 0; ii < len(photographer.Availabilities); ii++ {
			var availability = photographer.Availabilities[ii]
			fmt.Printf("\t\tstarts: %s\n", availability.Starts)
			fmt.Printf("\t\tends  : %s\n", availability.Ends)
		}

		fmt.Printf("\tbookings\n")
		for ii := 0; ii < len(photographer.Bookings); ii++ {
			var booking = photographer.Bookings[ii]
			fmt.Printf("\t\tstarts: %s\n", booking.Starts)
			fmt.Printf("\t\tends  : %s\n", booking.Ends)
		}
	}
}

func timeDiffMinutes(starts time.Time, ends time.Time) int {
	return int(ends.Sub(starts).Minutes())
}

func (photographers *Photographers) addAvailableTimeSlot(photographer Photographer, starts time.Time, durationInMinutes int) {
	photographers.Schedules = append(photographers.Schedules, Schedule{
		Photographer: photographer,
		TimeSlot: TimeSlot{
			Starts: starts,
			Ends:   starts.Add(time.Minute * time.Duration(durationInMinutes)),
		},
	})
}

func (photographers *Photographers) availableTimeSlotsForBooking(durationInMinutes int) []Schedule {
	for i := 0; i < len(photographers.Photographers); i++ {
		var photographer = photographers.Photographers[i]
		photog := Photographer{
			Id:   photographer.Id,
			Name: photographer.Name,
		}
	outter:
		for ii := 0; ii < len(photographer.Availabilities); ii++ {
			availability := photographer.Availabilities[ii]
			starts := availability.Starts
			ends := availability.Ends
			diff := timeDiffMinutes(starts, ends)
			if durationInMinutes > diff {
				continue
			}
			// assume bookings are sorted in chronological order
			for iii := 0; iii < len(photographer.Bookings); iii++ {
				booking := photographer.Bookings[iii]
				bookingStarts := booking.Starts
				bookingEnds := booking.Ends
				var prevEnds time.Time
				if bookingStarts.After(ends) && timeDiffMinutes(starts, ends) >= durationInMinutes {
					// add time slot
					photographers.addAvailableTimeSlot(photog, starts, durationInMinutes)
					break outter
				}
				if bookingStarts.Before(ends) {
					if iii == 0 {
						if timeDiffMinutes(starts, bookingStarts) >= durationInMinutes {
							// add time slot
							photographers.addAvailableTimeSlot(photog, starts, durationInMinutes)
							break outter
						}
					} else {
						if timeDiffMinutes(prevEnds, bookingStarts) >= durationInMinutes {
							// add time slot
							photographers.addAvailableTimeSlot(photog, prevEnds, durationInMinutes)
							break outter
						}
					}
				}
				// if last booking, compare booking end time with end time of available time slots
				if iii == len(photographer.Bookings)-1 {
					if timeDiffMinutes(bookingEnds, ends) >= durationInMinutes {
						photographers.addAvailableTimeSlot(photog, bookingEnds, durationInMinutes)
						break outter
					}
				}
				prevEnds = bookingEnds
			}
		}
	}
	return photographers.Schedules
}

func (photographers *Photographers) printSchedules(fname string, debug bool) (err error) {
	file, _ := json.MarshalIndent(photographers.Schedules, "", " ")
	err = ioutil.WriteFile(fname, file, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("### Successfully saved output in => %s\n", fname)

	if debug {
		fmt.Printf("Schedule:\n")
		for i := 0; i < len(photographers.Schedules); i++ {
			schedule := photographers.Schedules[i]
			photographer := schedule.Photographer
			timeSlot := schedule.TimeSlot
			fmt.Printf("Photographer\n")
			fmt.Printf("\tid  : %s\n", photographer.Id)
			fmt.Printf("\tName: %s\n", photographer.Name)
			fmt.Printf("Time Slot\n")
			fmt.Printf("\tstarts: %s\n", timeSlot.Starts)
			fmt.Printf("\tends  : %s\n", timeSlot.Ends)

		}
	}
	return
}

func main() {
	input := flag.String("input", "", "Input json file containing photographer availability and bookings")
	debug := flag.Bool("debug", false, "Print debug output")
	flag.Parse()
	if *input == "" {
		fmt.Println("Please provide input file. You can find usage by running \" ./schedule --help\"")
		os.Exit(1)
	}

	photographers := Photographers{}
	err := photographers.readInput(*input)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}
	if *debug {
		photographers.printInput()
	}

	photographers.availableTimeSlotsForBooking(90)
	err = photographers.printSchedules(fmt.Sprintf("%s.output", *input), *debug)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}
}
