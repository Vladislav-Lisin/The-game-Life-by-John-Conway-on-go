package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

var (
	square = [][]int{
		{0, 1},
		{1, 0},
		{0, 0},
		{1, 1},
	}

	glider = [][]int{
		{1, 0},
		{2, 1},
		{0, 2},
		{1, 2},
		{2, 2},
	}

	blinker = [][]int{
		{0, 0},
		{1, 0},
		{2, 0},
	}

	rPentomino = [][]int{
		{1, 0},
		{2, 0},
		{0, 1},
		{1, 1},
		{1, 2},
	}

	acorn = [][]int{
		{1, 0},
		{3, 1},
		{0, 2},
		{1, 2},
		{4, 2},
		{5, 2},
		{6, 2},
	}
)

const alive = "■"

func main() {
	var n, m int
	var shapes []int

	//p := runtime.NumCPU()
	// var lifeTime int = 1

	fmt.Println("Введите параметры NxM для поля игры.")
	fmt.Print("N:")
	fmt.Scan(&n)
	fmt.Print("M:")
	fmt.Scan(&m)

	arrA := make([][]string, n)
	arrB := make([][]string, n)

	for i := 0; i < n; i++ {
		arrA[i] = make([]string, m)
		arrB[i] = make([]string, m)
		for j := range arrA[i] {
			arrA[i][j] = "*"
			arrB[i][j] = "*"
		}
	}
	choosShape(&shapes)
	fillingMatrix(&arrA, shapes)
	printArr(&arrA)

	cur, next := arrA, arrB
	workers := runtime.NumCPU()
	jobs := make(chan job)
	wg := &sync.WaitGroup{}

	for i := 0; i < workers; i++ {
		go worker(jobs, wg) // пул запускается один раз
	}

	fmt.Print("\033[2J") // очистить экран один раз
	gen := 0
	for gen < 200 {
		step := (n + workers - 1) / workers // размер полосы
		for from := 0; from < n; from += step {
			to := from + step
			if to > n {
				to = n
			}
			wg.Add(1) // до отправки задачи
			jobs <- job{from, to, cur, next}
		}
		wg.Wait()
		cur, next = next, cur

		fmt.Print("\033[H")
		printArr(&cur)
		time.Sleep(100 * time.Millisecond)
		gen++
	}
	close(jobs)
	fmt.Printf("\nПрожито итераций: %d\n", gen)
}

func choosShape(shapes *[]int) {
	var f int
	fmt.Print("/nВведите номера фигур изначально живых:\n")
	fmt.Print("1. Квадрат\n2. Глайдер\n3. Блинкер\n4. R-пентамино\n5. Жёлудь\n0. Конец выбора фигур\n")
	for {
		fmt.Scan(&f)
		if f == 0 {
			break
		}
		if f >= 1 && f <= 5 {
			*shapes = append(*shapes, f)
		} else {
			fmt.Println("Ошибка: введите число от 0 до 5")
		}
		fmt.Println("Фигура добавлена! Выберите ещё или выйдите из цикла")
	}

}

func fillingMatrix(arr *[][]string, shapesArr []int) {
	m := *arr
	if len(m) == 0 {
		return
	}

	for _, v := range shapesArr {
		var shape [][]int

		switch v {
		case 1:
			shape = square
		case 2:
			shape = glider
		case 3:
			shape = blinker
		case 4:
			shape = rPentomino
		case 5:
			shape = acorn
		case 0:
			return
		default:
			continue
		}

		placeShape(m, shape)
	}
}

func placeShape(m [][]string, shape [][]int) {
	height, width := len(m), len(m[0])

	// размер фигуры, чтобы случайная точка не вывела её за границы поля
	maxX, maxY := 0, 0
	for _, c := range shape {
		if c[0] > maxX {
			maxX = c[0]
		}
		if c[1] > maxY {
			maxY = c[1]
		}
	}

	if width <= maxX || height <= maxY {
		return // фигура не помещается в поле
	}

	x0 := rand.Intn(width - maxX)
	y0 := rand.Intn(height - maxY)

	for _, c := range shape {
		m[y0+c[1]][x0+c[0]] = alive
	}
}

func printArr(arr *[][]string) {
	for _, row := range *arr {
		for _, val := range row {
			fmt.Printf("%s ", val)
		}
		fmt.Println()
	}
}

type job struct {
	from, to  int // диапазон строк [from, to)
	cur, next [][]string
}

func worker(jobs <-chan job, wg *sync.WaitGroup) {
	for j := range jobs {
		computeRows(j)
		wg.Done()
	}
}

func computeRows(j job) {
	height, width := len(j.cur), len(j.cur[0])

	for y := j.from; y < j.to; y++ {
		for x := 0; x < width; x++ {
			count := 0
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}
					// тороидальное поле: края замыкаются
					ny := (y + dy + height) % height
					nx := (x + dx + width) % width
					if j.cur[ny][nx] == alive {
						count++
					}
				}
			}

			isAlive := j.cur[y][x] == alive
			if (isAlive && (count == 2 || count == 3)) || (!isAlive && count == 3) {
				j.next[y][x] = alive
			} else {
				j.next[y][x] = "*"
			}
		}
	}
}
