package tests

import (
	"math"
	"os"
)

type TestFile struct {
	Path           string
	PartitionScore int64
}

func Split(files []string, index int64, total int64) []string {
	testFiles := []TestFile{}

	for i := 0; i < len(files); i++ {
		testFile := pathToTestFile(files[i])
		if testFile.PartitionScore > 0 {
			testFiles = append(testFiles, testFile)
		}
	}

	testFiles = partition(testFiles, total)[index]
	return mapTestFilesToPaths(testFiles)
}

func partition(testFiles []TestFile, total int64) [][]TestFile {
	if total <= int64(1) {
		return [][]TestFile{testFiles}
	}

	if total >= int64(len(testFiles)) {
		t := [][]TestFile{}
		for i := 0; i <= len(testFiles); i++ {
			t = append(t, []TestFile{testFiles[i]})
		}

		return t
	}

	// create a list of indexes to partition between, using the index on the
	// left of the partition to indicate where to partition
	// to start, roughly partition the array into equal groups of len(a)/k (note
	// that the last group may be a different size)
	partitionBetween := []int64{}
	for i := int64(0); i < total-1; i++ {
		partitionBetween = append(partitionBetween, (i+1)*int64(len(testFiles))/total)
	}

	// the ideal size for all partitions is the total height of the list divided
	// by the number of paritions
	var bestScore int64 = -1
	averageHeight := sum(testFiles) / total
	bestPartitions := [][]TestFile{}
	count := 0
	noImprovementsCount := 0

	for true {
		partitions := [][]TestFile{}
		var index int64 = 0

		for _, div := range partitionBetween {
			// create partitions based on partitionBetween
			partitions = append(partitions, testFiles[index:div])
			index = div
		}

		// append the last partition, which runs from the last partition divider
		// to the end of the list
		partitions = append(partitions, testFiles[index:])

		// evaluate the partitioning
		var worstHeightDiff int64
		var worstPartitionIndex int64 = -1

		for index, p := range partitions {
			// compare the partition height to the average partition height
			heightDiff := averageHeight - sum(p)
			// if it's the worst partition we've seen, update the variables that
			// track that
			if int64(math.Abs(float64(heightDiff))) > int64(math.Abs(float64(worstHeightDiff))) {
				worstHeightDiff = heightDiff
				worstPartitionIndex = int64(index)
			}
		}
		// if the worst partition from this run is still better than anything
		// we saw in previous iterations, update our best-ever variables
		if bestScore == -1 || int64(math.Abs(float64(worstHeightDiff))) < bestScore {
			bestScore = int64(math.Abs(float64(worstHeightDiff)))
			bestPartitions = partitions
			noImprovementsCount = 0
		} else {
			noImprovementsCount += 1
		}

		// decide if we're done: if all our partition heights are ideal, or if
		// we haven't seen improvement in >5 iterations, or we've tried 100
		// different partitionings
		// the criteria to exit are important for getting a good result with
		// complex data, and changing them is a good way to experiment with getting
		// improved results
		if worstHeightDiff == 0 || noImprovementsCount > 5 || count > 100 {
			return bestPartitions
		}
		count += 1

		// adjust the partitioning of the worst partition to move it closer to the
		// ideal size. the overall goal is to take the worst partition and adjust
		// its size to try and make its height closer to the ideal. generally, if
		// the worst partition is too big, we want to shrink the worst partition
		// by moving one of its ends into the smaller of the two neighboring
		// partitions. if the worst partition is too small, we want to grow the
		// partition by expanding the partition towards the larger of the two
		// neighboring partitions
		if worstPartitionIndex == 0 { // the worst partition is the first one
			if worstHeightDiff < 0 {
				partitionBetween[0] -= 1 // partition too big, so make it smaller
			} else {
				partitionBetween[0] += 1 // partition too small, so make it bigger
			}
		} else if worstPartitionIndex == int64(len(partitions))-1 { // the worst partition is the last one
			if worstHeightDiff < 0 {
				partitionBetween[len(partitionBetween)-1] += 1 // partition too small, so make it bigger
			} else {
				partitionBetween[len(partitionBetween)-1] -= 1 // partition too big, so make it smaller
			}
		} else { // the worst partition is in the middle somewhere
			leftBound := worstPartitionIndex - 1 // the divider before the partition
			rightBound := worstPartitionIndex    // the divider after the partition
			if worstHeightDiff < 0 {             // partition too big, so make it smaller
				// the partition on the left is bigger than the one on the right, so make the one on the right bigger
				if sum(partitions[worstPartitionIndex-1]) > sum(partitions[worstPartitionIndex+1]) {
					partitionBetween[rightBound] -= 1
					// the partition on the left is smaller than the one on the right, so make the one on the left bigger
				} else {
					partitionBetween[leftBound] += 1
				}
			} else { // partition too small, make it bigger
				// the partition on the left is bigger than the one on the right, so make the one on the left smaller
				if sum(partitions[worstPartitionIndex-1]) > sum(partitions[worstPartitionIndex+1]) {
					partitionBetween[leftBound] -= 1
				} else {
					// the partition on the left is smaller than the one on the right, so make the one on the right smaller
					partitionBetween[rightBound] += 1
				}
			}
		}
	}

	return bestPartitions
}

func mapTestFilesToPaths(testFiles []TestFile) []string {
	paths := []string{}

	for i := 0; i < len(testFiles); i++ {
		paths = append(paths, testFiles[i].Path)
	}

	return paths
}

func pathToTestFile(path string) TestFile {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return TestFile{}
	} else {
		return TestFile{Path: path, PartitionScore: fileInfo.Size()}
	}
}

func sum(input []TestFile) int64 {
	scores := int64(0)
	for _, tf := range input {
		scores += tf.PartitionScore
	}

	return scores
}
