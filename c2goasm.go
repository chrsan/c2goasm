/*
 * Minio Cloud Storage, (C) 2017 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package c2goasm

import (
	"fmt"
	"strings"
)

type Result struct {
	Sub Subroutine
	ASM []string
}

func Process(assembly []string, goCompanionFile string, stackSizes map[string]uint) ([]Result, error) {

	// Split out the assembly source into subroutines
	subroutines := segmentSource(assembly)
	tables := segmentConstTables(assembly)

	var result []Result

	// Iterate over all subroutines
	for _, sub := range subroutines {
		var r Result
		golangArgs, golangReturns := parseCompanionFile(goCompanionFile, sub.Name)
		stackArgs := argumentsOnStack(sub.Body)
		if len(golangArgs) > 6 && len(golangArgs)-6 < stackArgs.Number {
			panic(fmt.Sprintf("Found too few arguments on stack (%d) but needed %d", len(golangArgs)-6, stackArgs.Number))
		}

		// Check for constants table
		if table := getCorrespondingTable(sub.Body, tables); table.isPresent() {

			// Output constants table
			r.ASM = append(r.ASM, strings.Split(table.Constants, "\n")...)
			r.ASM = append(r.ASM, "") // append empty line

			sub.Table = table
		}

		// Create object to get offsets for stack pointer
		sub.Stack = NewStack(sub.Epilogue, len(golangArgs), scanBodyForCalls(sub, stackSizes))

		// Write header for subroutine in go assembly
		r.ASM = append(r.ASM, writeGoasmPrologue(sub, golangArgs, golangReturns)...)

		// Write body of code
		assembly, err := writeGoasmBody(sub, stackArgs, golangArgs, golangReturns)
		if err != nil {
			panic(fmt.Sprintf("writeGoasmBody: %v", err))
		}
		r.ASM = append(r.ASM, assembly...)
		r.Sub = sub
		result = append(result, r)
	}

	return result, nil
}
