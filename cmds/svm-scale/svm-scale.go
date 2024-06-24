/*
** Copyright 2024 R. Mera A.
**
** Licensed under the Apache License, Version 2.0 (the "License");
** you may not use this file except in compliance with the License.
** You may obtain a copy of the License at
**
** http ://www.apache.org/licenses/LICENSE-2.0
**
** Unless required by applicable law or agreed to in writing, software
** distributed under the License is distributed on an "AS IS" BASIS,
** WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
** See the License for the specific language governing permissions and
** limitations under the License.
**
** @author: R. Mera
 */
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	libSvm "github.com/rmera/libsvm-go"
)

func main() {
	param := libSvm.NewParameter()                              // create a parameter type
	scalefile, scfilesave, trainFile, lu := parseOptions(param) // parse command-line flags for SVM parameter

	prob, err := libSvm.NewProblem(trainFile, param) // create a problem type from the train file and the parameter
	if err != nil {
		fmt.Fprint(os.Stderr, "Fail to create a libSvm.Problem: ", err)
		os.Exit(1)
	}
	var minmax [][2]float64

	if scalefile != "" {
		//note that the file has priority over the "lu" options.
		minmax, lu, err = libSvm.ReadRangesFromFile(scalefile)
	}
	newminmax, err := prob.Scale(minmax, lu...)
	if err != nil {
		log.Printf("Encountered error while scaling: %v", err)
	}

	for prob.Begin(); !prob.Done(); prob.Next() {
		fmt.Println(prob.String())
	}
	if scfilesave != "" {
		err := libSvm.WriteRanges2File(scfilesave, newminmax, lu)
		if err != nil {
			log.Printf("Couldn't write range file: %v", err)
		}
	}

}

func usage() {
	fmt.Print(
		"Usage: svm-scale [options] training_set_file \n",
		"options:\n",
		"-l: Scaling lower limit (default -1)",
		"-u: Scaling upper limit (default 1)",
		"-r file: Restore scaling parameters from file. Overwrites the -l and -u options\n",
		"-y: y_lower: y scaling limits. Only for compatibility, not actually used.",
		"-s file: Save scaling parameters to file\n")
}

func parseOptions(param *libSvm.Parameter) (string, string, string, []float64) {
	var scalefile, scalefilesave, w string
	var l, u float64
	flag.StringVar(&w, "y", "", "") //not used.
	flag.StringVar(&scalefile, "r", "", "")
	flag.StringVar(&scalefilesave, "s", "", "")
	flag.Float64Var(&l, "l", -1, "")
	flag.Float64Var(&u, "u", 1, "")
	flag.Usage = usage
	flag.Parse()
	if len(flag.Args()) < 1 {
		usage()
		os.Exit(1)
	}
	return scalefile, scalefilesave, flag.Args()[0], []float64{l, u}
}
