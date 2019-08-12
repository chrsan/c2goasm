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

func Process(assembly []string, goCompanionFile string, stackSizes map[string]uint) ([]string, error) {

	// Split out the assembly source into subroutines
	subroutines := segmentSource(assembly)
	tables := segmentConstTables(assembly)

	var result []string

	// Iterate over all subroutines
	for isubroutine, sub := range subroutines {

		golangArgs, golangReturns := parseCompanionFile(goCompanionFile, sub.name)
		stackArgs := argumentsOnStack(sub.body)
		if len(golangArgs) > 6 && len(golangArgs)-6 < stackArgs.Number {
			panic(fmt.Sprintf("Found too few arguments on stack (%d) but needed %d", len(golangArgs)-6, stackArgs.Number))
		}

		// Check for constants table
		if table := getCorrespondingTable(sub.body, tables); table.isPresent() {

			// Output constants table
			result = append(result, strings.Split(table.Constants, "\n")...)
			result = append(result, "") // append empty line

			sub.table = table
		}

		// Create object to get offsets for stack pointer
		stack := NewStack(sub.epilogue, len(golangArgs), scanBodyForCalls(sub, stackSizes))

		// Write header for subroutine in go assembly
		result = append(result, writeGoasmPrologue(sub, stack, golangArgs, golangReturns)...)

		// Write body of code
		assembly, err := writeGoasmBody(sub, stack, stackArgs, golangArgs, golangReturns)
		if err != nil {
			panic(fmt.Sprintf("writeGoasmBody: %v", err))
		}
		result = append(result, assembly...)

		if isubroutine < len(subroutines)-1 {
			// Empty lines before next subroutine
			result = append(result, "\n", "\n")
		}
	}

	return result, nil
}
