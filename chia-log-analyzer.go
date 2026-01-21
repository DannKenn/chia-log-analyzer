package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

var debuglogFile = flag.String("log", "./debug.log", "path to debug.log")
var widgetLastPlots, widgetFoundProofs, widgetLastFarmingTime, widgetTotalFarmingPlotsNumber, widgetLog *widgets.Paragraph
var widgetBarChart *widgets.BarChart
var widgetBarChart2 *widgets.Plot
var widgetSparklines *widgets.Sparkline
var widgetSparklinesGroup *widgets.SparklineGroup
var widgetOverallHealthPercent *widgets.Paragraph
var lastRow string = ""
var totalFarmingAttempt, positiveFarmingAttempt, foundProofs int = 0, 0, 0
var farmingTime, totalPlots string = "0", "0"
var minFarmingTime, maxFarmingTime float64 = 999999.0, 0.0
var allFarmingTimes []float64
var lastLogFileSize int64 = 0
var healthData = make(map[string]float64)

type stackStruct struct { lines []string; count int }
type stackStructFloats struct { values []float64; count int }
type poolInfoStruct struct { name string; partialsCount int }

func (s *stackStruct) push(l string) { s.lines = append(s.lines, l); if len(s.lines) > s.count { s.lines = s.lines[1:] } }
func (s *stackStructFloats) push(v float64) { s.values = append(s.values, v); if len(s.values) > s.count { s.values = s.values[1:] } }

var lastParsedLinesStack = stackStruct{count: 5}
var lastFarmStack = stackStructFloats{count: 29}
var lastFarmingTimesStack = stackStructFloats{count: 113}
var poolInfo = poolInfoStruct{partialsCount: 0}

func main() {
	detectLogFileLocation()
	if err := ui.Init(); err != nil { panic(err) }; defer ui.Close()
	setupWidgets()
	go loopReadFile()
	for e := range ui.PollEvents() { if e.ID == "q" || e.ID == "<C-c>" { return } }
}

func parseLines(lines []string) {
	// 2.5.7 FORMATINA OZEL REGEX: 'Found 1 V1 proofs' ve 'Found 0 proofs' kalıplarını yakalar.
	rePlots := regexp.MustCompile(`(\d+)\s+plots\s+were\s+eligible.*Found\s+(\d+)\s+(?:V\d+\s+)?proofs.*Time:\s+([0-9\.]+)\s+s.*Total\s+(\d+)\s+plots`)
	rePool := regexp.MustCompile(`Submitting\s+partial.*to\s+(https?://\S+)`)
	reTime := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})`)

	start := false
	for _, s := range lines {
		if s == "" { continue }
		if !start { if lastRow == "" || lastRow == s { start = true }; continue }
		lastRow = s
		if rePlots.MatchString(s) {
			lastParsedLinesStack.push(s)
			if t := reTime.FindString(s); t != "" { healthData[t[0:15]]++ }
			m := rePlots.FindStringSubmatch(s)
			if len(m) >= 5 {
				eligible, _ := strconv.Atoi(m[1])
				proofs, _ := strconv.Atoi(m[2])
				fTime, _ := strconv.ParseFloat(m[3], 64)
				foundProofs += proofs
				farmingTime, totalPlots = m[3], m[4]
				if eligible > 0 { positiveFarmingAttempt += eligible }
				totalFarmingAttempt++
				lastFarmStack.push(float64(eligible))
				lastFarmingTimesStack.push(fTime)
				allFarmingTimes = append(allFarmingTimes, fTime)
				if fTime < minFarmingTime { minFarmingTime = fTime }
				if fTime > maxFarmingTime { maxFarmingTime = fTime }
			}
		}
		if rePool.MatchString(s) {
			m := rePool.FindStringSubmatch(s)
			if len(m) > 1 { poolInfo.name, poolInfo.partialsCount = m[1], poolInfo.partialsCount+1 }
		}
	}
	renderAll()
}

func setupWidgets() {
	h := 3
	widgetLastPlots = widgets.NewParagraph(); widgetLastPlots.Title = "Plots"; widgetLastPlots.SetRect(0, 0, 10, h)
	widgetFoundProofs = widgets.NewParagraph(); widgetFoundProofs.Title = "Proofs/Pool"; widgetFoundProofs.SetRect(10, 0, 45, h)
	widgetTotalFarmingPlotsNumber = widgets.NewParagraph(); widgetTotalFarmingPlotsNumber.Title = "Attempts"; widgetTotalFarmingPlotsNumber.SetRect(45, 0, 70, h)
	widgetLastFarmingTime = widgets.NewParagraph(); widgetLastFarmingTime.Title = "Times (L/Min/Avg/Max)"; widgetLastFarmingTime.SetRect(70, 0, 105, h)
	widgetOverallHealthPercent = widgets.NewParagraph(); widgetOverallHealthPercent.Title = "Health"; widgetOverallHealthPercent.SetRect(105, 0, 119, h)
	widgetLog = widgets.NewParagraph(); widgetLog.Title = "Last Activity"; widgetLog.SetRect(0, 3, 119, 12)
	widgetBarChart = widgets.NewBarChart(); widgetBarChart.Title = "Eligible Plots"; widgetBarChart.SetRect(0, 12, 119, 22); widgetBarChart.BarWidth = 3; widgetBarChart.BarColors = []ui.Color{ui.ColorGreen}
	widgetBarChart2 = widgets.NewPlot(); widgetBarChart2.Title = "Latency (s)"; widgetBarChart2.SetRect(0, 22, 119, 32); widgetBarChart2.Data = make([][]float64, 1); widgetBarChart2.LineColors[0] = ui.ColorRed
	widgetSparklines = widgets.NewSparkline(); widgetSparklines.LineColor = ui.ColorBlue
	widgetSparklinesGroup = widgets.NewSparklineGroup(widgetSparklines); widgetSparklinesGroup.Title = "Network Health"; widgetSparklinesGroup.SetRect(0, 32, 119, 40)
}

func renderAll() {
	p := 0.0; if totalFarmingAttempt > 0 { p = float64(positiveFarmingAttempt) / float64(totalFarmingAttempt) * 100 }
	widgetTotalFarmingPlotsNumber.Text = fmt.Sprintf("%d/%d(%.1f%%)", positiveFarmingAttempt, totalFarmingAttempt, p)
	widgetLastPlots.Text = totalPlots
	pT := fmt.Sprintf("%d/%d", foundProofs, poolInfo.partialsCount); if poolInfo.partialsCount > 0 { pT += " (" + poolInfo.name + ")" }; widgetFoundProofs.Text = pT
	avg := 0.0; if len(allFarmingTimes) > 0 { sum := 0.0; for _, v := range allFarmingTimes { sum += v }; avg = sum / float64(len(allFarmingTimes)) }
	widgetLastFarmingTime.Text = fmt.Sprintf("%ss/%.3fs/%.3fs/%.3fs", farmingTime, minFarmingTime, avg, maxFarmingTime)
	if len(healthData) >= 3 {
		v := sortMap(healthData); v = v[1 : len(v)-1]; hP := (sumFloats(v) / float64(len(v))) / 64 * 100
		widgetOverallHealthPercent.Text = fmt.Sprintf("%.2f%%", hP)
		if hP > 90 { widgetOverallHealthPercent.TextStyle.Fg = ui.ColorGreen } else { widgetOverallHealthPercent.TextStyle.Fg = ui.ColorRed }
	}
	var sb strings.Builder
	for _, l := range lastParsedLinesStack.lines {
		msg := l; if i := strings.LastIndex(l, "INFO"); i != -1 { msg = l[i+4:] }; sb.WriteString(strings.TrimSpace(msg) + "\n")
	}
	widgetLog.Text = sb.String()
	widgetBarChart.Data = lastFarmStack.values
	widgetBarChart2.Data[0] = lastFarmingTimesStack.values
	vS := sortMap(healthData); if len(vS) > 117 { vS = vS[len(vS)-117:] }; widgetSparklines.Data = vS
	ui.Render(widgetLastPlots, widgetFoundProofs, widgetTotalFarmingPlotsNumber, widgetLastFarmingTime, widgetOverallHealthPercent, widgetLog, widgetBarChart, widgetBarChart2, widgetSparklinesGroup)
}

func detectLogFileLocation() {
	if _, err := os.Stat(*debuglogFile); err == nil { return }
	u, _ := user.Current(); d := filepath.Join(u.HomeDir, ".chia", "mainnet", "log", "debug.log")
	debuglogFile = &d
}

func loopReadFile() {
	readFullFile(*debuglogFile); setLastLogFileSize(*debuglogFile)
	for range time.Tick(5 * time.Second) {
		s, _ := getFileSize(*debuglogFile); if s == 0 || s == lastLogFileSize { continue }
		if s < lastLogFileSize { readFullFile(*debuglogFile) } else { readFile(*debuglogFile) }; setLastLogFileSize(*debuglogFile)
	}
}

func setLastLogFileSize(p string) { lastLogFileSize, _ = getFileSize(p) }
func getFileSize(p string) (int64, error) { fi, err := os.Stat(p); if err != nil { return 0, err }; return fi.Size(), nil }
func readFullFile(f string) { b, _ := ioutil.ReadFile(f); lines := strings.Split(string(b), "\n"); parseLines(lines) }
func readFile(f string) {
	file, _ := os.Open(f); defer file.Close(); stat, _ := file.Stat(); start := int64(0)
	if stat.Size() > 16384 { start = stat.Size() - 16384 }; buf := make([]byte, 16384); n, _ := file.ReadAt(buf, start)
	lines := strings.Split(string(buf[:n]), "\n"); parseLines(lines)
}

func sortMap(m map[string]float64) []float64 {
	keys := make([]string, 0, len(m)); for k := range m { keys = append(keys, k) }; sort.Strings(keys)
	v := make([]float64, 0, len(m)); for _, k := range keys { v = append(v, m[k]) }; return v
}

func sumFloats(i []float64) float64 { s := 0.0; for _, v := range i { s += v }; return s }