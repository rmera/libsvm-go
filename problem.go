/*
** Copyright 2014 Edward Walker
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
** Description: Describes problem, i.e. label/vector set
** @author: Ed Walker
 */
package libSvm

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type snode struct {
	index int     // dimension (-1 indicates end of SV)
	value float64 // coeff
}

type Problem struct {
	l      int       // #SVs
	y      []float64 // labels
	x      []int     // starting indices in xSpace defining SVs
	xSpace []snode   // SV coeffs
	i      int       // counter for iterator
}

func NewProblem(file string, param *Parameter) (*Problem, error) {
	prob := &Problem{l: 0, i: 0}
	err := prob.Read(file, param)
	return prob, err
}

func scale(v, l, u float64, mm [2]float64) float64 {
	return l + (u-l)*(v-mm[0])/(mm[1]-mm[0])

}

//rmera: Not completely sure of this one.
//It will absolutely fail if not all vector have these same amount of components.
func (problem *Problem) Features() int {
	return problem.x[1] - problem.x[0]
}

//rmera: Scales the whole problem so that, for each feature, all its values run between lowandup[0]
//and lowandup[1] (-1 and 1 by default, respectively). minmax is a with as many elements as
//features, where each element is a 2-array with the min and max for the respective feature.
//if minmax is nil, it will be obtained by the function. If it is given but doesn't have enough
//elements, Scale will return an error.
//sorry about using "p" instead of "problem" for the receiver :(
func (p *Problem) Scale(minmax [][2]float64, lowandup ...float64) ([][2]float64, error) {
	u := 1.0
	l := -1.0
	if len(lowandup) > 0 {
		l = lowandup[0]
	}
	if len(lowandup) > 1 {
		u = lowandup[1]
	}
	if minmax == nil {
		minmax = make([][2]float64, 0, len(p.x))
		for j, v := range p.x {
			for i := v; p.xSpace[i].index != -1; i++ {
				if j == 0 {
					minmax = append(minmax, [2]float64{p.xSpace[i].value, p.xSpace[i].value})
				}
				val := p.xSpace[i].value
				if minmax[i-v][0] > val {
					minmax[i-v][0] = val
				}
				if minmax[i-v][1] < val {
					minmax[i][1] = val
				}

			}
		}

	}
	_, m := p.GetLine()
	features := len(m)
	if len(minmax) < features {
		return nil, fmt.Errorf("Not enough minmax values supplied. Got: %d, need: %d", len(minmax), p.Features())
	}
	for _, v := range p.x {
		for i := v; p.xSpace[i].index != -1; i++ {
			val := p.xSpace[i].value
			if len(minmax) <= i-v {
				//This should mean that vectors don't have all the same dimensions.
				panic(fmt.Sprintf("Bug! Code wrongly predicts the number of minmax needed. Prediction: %d", features))
			}
			p.xSpace[i].value = scale(val, l, u, minmax[i-v])
		}
	}
	return minmax, nil
}

func (problem *Problem) Read(file string, param *Parameter) error { // reads the problem from the specified file
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("Fail to open file %s\n", file)
	}

	defer f.Close() // close f on method return

	problem.y = nil
	problem.x = nil
	problem.xSpace = nil

	reader := bufio.NewReader(f)
	var max_idx int = 0
	var l int = 0

	for {
		line, err := readline(reader)
		if err != nil {
			break
		}
		problem.x = append(problem.x, len(problem.xSpace))

		lineSansComments := strings.Split(line, "#") // remove any comments

		tokens := strings.Fields(lineSansComments[0]) // get all the word tokens (seperated by white spaces)
		if label, err := strconv.ParseFloat(tokens[0], 64); err == nil {
			problem.y = append(problem.y, label)
		} else {
			return fmt.Errorf("Fail to parse label\n")
		}

		space := tokens[1:]
		for _, w := range space {
			if len(w) > 0 {
				node := strings.Split(w, ":")
				if len(node) > 1 {
					var index int
					var value float64
					if index, err = strconv.Atoi(node[0]); err != nil {
						return fmt.Errorf("Fail to parse index from token %v\n", w)
					}
					if value, err = strconv.ParseFloat(node[1], 64); err != nil {
						return fmt.Errorf("Fail to parse value from token %v\n", w)
					}
					problem.xSpace = append(problem.xSpace, snode{index: index, value: value})
					if index > max_idx {
						max_idx = index
					}

				}
			}
		}

		problem.xSpace = append(problem.xSpace, snode{index: -1})
		l++
	}
	problem.l = l

	if param.Gamma == 0 && max_idx > 0 {
		param.Gamma = 1.0 / float64(max_idx)
	}

	return nil
}

/**
 * Initialize the start of iterating through the labels and vectors in the problem set
 */
func (problem *Problem) Begin() {
	problem.i = 0
}

/**
 * Finished iterating through all the labels and vectors in the problem set
 */
func (problem *Problem) Done() bool {
	if problem.i >= problem.l {
		return true
	}
	return false
}

/**
 * Move to the next label and vector in the problem set
 */
func (problem *Problem) Next() {
	problem.i++
	return
}

/**
 * Return one label and vector from the problem set
 * @return y label
 * @return x vector (map of dimension/value)
 */
func (problem *Problem) GetLine() (y float64, x map[int]float64) {
	y = problem.y[problem.i]
	idx := problem.x[problem.i]
	x = SnodeToMap(problem.xSpace[idx:])
	return // y, x
}

/**
 * Returns number of label and vectors in the problem set
 * @return problem set size
 */
func (problem *Problem) ProblemSize() int {
	return problem.l
}

//rmera return the problem as a slice of string
//where each string represents a vector. The string
//is formated in the libSVM format:
//label index1:value1 index2:value2 (...)
func (problem *Problem) Strings() []string {
	ret := make([]string, 0, problem.l)
	for j := 0; j < problem.l; j++ {
		y, m := problem.GetLine()
		keys := make([]int, len(m))
		i := 0
		for k := range m {
			keys[i] = k
			i++
		}
		sort.Ints(keys)
		fields := make([]string, 1, len(keys))
		fields[0] = fmt.Sprintf("%5.3f", y)
		for _, v := range keys {
			s := fmt.Sprintf("%d:%3.5f", v, m[v])
			fields = append(fields, s)
		}
		ret = append(ret, strings.Join(fields, " "))
	}
	return ret
}
