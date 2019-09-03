/**
 * File  : outfile.go
 * Author: Reinaldo Astudillo <r.a.astudillo@tudelft.nl>
 */
package outfile

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type IterInfo struct {
	Number          int
	PreparingTime   time.Duration
	SolvingTime     time.Duration
	Eigenvalues     []float64
	Objective       float64
	VolumeConstrain float64
	DesignChange    float64
}

func (iter *IterInfo) Parse(data []string) {
	for _, line := range data {
		if !strings.Contains(line, ":") {
			continue
		}
		val := getDataAfterColon(line)
		if strings.HasPrefix(line, "Iteration:") {
			n, err := strconv.Atoi(val)
			if err != nil {
				log.Println("invalid iter number")
				n = -1
			}
			iter.Number = n
			continue
		}
		if strings.HasPrefix(line, "Objective:") {
			v, err := strconv.ParseFloat(val, 64)
			if err != nil {
				log.Println("invalid Objective number")
				v = -1.0
			}
			iter.Objective = v
			continue
		}
		if strings.HasPrefix(line, "Eigenvalue") {
			v, err := strconv.ParseFloat(val, 64)
			if err != nil {
				log.Println("invalid Objective number")
				v = -1.0
			}
			iter.Eigenvalues = append(iter.Eigenvalues, v)
			continue
		}
		if strings.HasPrefix(line, "Volume constraint:") {
			v, err := strconv.ParseFloat(val, 64)
			if err != nil {
				log.Println("invalid Objective number")
				v = -1.0
			}
			iter.VolumeConstrain = v
			continue
		}
		if strings.HasPrefix(line, "Design change:") {
			v, err := strconv.ParseFloat(val, 64)
			if err != nil {
				log.Println("invalid Objective number")
				v = -1.0
			}
			iter.DesignChange = v
			continue
		}
		if strings.HasPrefix(line, "Preparing:") {
			iter.PreparingTime, _ = getTime(val)
			continue
		}
		if strings.HasPrefix(line, "Solving:") {
			iter.SolvingTime, _ = getTime(val)
			continue
		}
	}
}

func (iter IterInfo) String() string {
	b, _ := json.MarshalIndent(iter, "", "  ")
	return string(b)
}

type FullInfo struct {
	Iterations []IterInfo
}

func (info FullInfo) reduceIterations(key func(IterInfo) time.Duration, op func(time.Duration, time.Duration) time.Duration) time.Duration {
	var result time.Duration
	for _, it := range info.Iterations {
		result = op(result, key(it))
	}
	return result
}

func timeSolPrep(i IterInfo) time.Duration {
	return i.SolvingTime + i.PreparingTime
}

func timeSol(i IterInfo) time.Duration {
	return i.SolvingTime
}

func timePrep(i IterInfo) time.Duration {
	return i.PreparingTime
}

func sum(a, b time.Duration) time.Duration {
	return a + b
}

func max(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func (info FullInfo) TotalTimePreparing() time.Duration {
	return info.reduceIterations(timePrep, sum)
}

func (info FullInfo) TotalTimeSolving() time.Duration {
	return info.reduceIterations(timeSol, sum)
}

func (info FullInfo) TotalTime() time.Duration {
	return info.reduceIterations(timeSolPrep, sum)
}

func (info FullInfo) AvgTime() time.Duration {
	return info.TotalTime() / time.Duration(info.NumberOfIter())
}

func (info FullInfo) NumberOfIter() int {
	return len(info.Iterations)
}

func (info *FullInfo) Load(f string) error {
	data, err := GetRawFile(f)
	if err != nil {
		return err
	}
	b, e := getLimIter(data)
	info.Iterations = make([]IterInfo, 0, len(b))
	for i := range b {
		begin, end := b[i], e[i]
		var iter IterInfo
		iter.Parse(data[begin : end+1])
		info.Iterations = append(info.Iterations, iter)
	}
	return nil
}

func GetRaw(r io.Reader) ([]string, error) {
	sc := bufio.NewScanner(r)
	data := make([]string, 0, 100)
	for sc.Scan() {
		line := sc.Text()
		data = append(data, line)
	}
	return data, sc.Err()
}

func GetRawFile(f string) ([]string, error) {
	file, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return GetRaw(file)
}

func getLimIter(data []string) ([]int, []int) {
	var b []int
	var e []int
	for ind, line := range data {
		if strings.HasPrefix(line, "Iteration: ") {
			b = append(b, ind)
			e = append(e, ind)
		} else if strings.HasPrefix(line, "Design change: ") {
			e[len(e)-1] = ind
		}
	}
	return b, e
}

func getDataAfterColon(line string) string {
	return strings.TrimSpace(line[strings.LastIndex(line, ":")+1:])
}

func getTime(t string) (time.Duration, error) {
	text := strings.ReplaceAll(strings.ReplaceAll(t, " seconds ", "s"),
		" milliseconds", "ms")
	return time.ParseDuration(text)
}
