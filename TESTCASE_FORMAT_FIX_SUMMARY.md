# Testcase Input Format Fix Summary

## Problem Identified

The AI Build Test feature was generating testcases in Python variable assignment format:
```
arr = [-1, -2, -3, -4]
```

But the judge executor expects standard competitive programming stdin format:
```
4
-1 -2 -3 -4
```

This caused **runtime errors on ALL testcases** because:
1. Python code does `n = int(input())` expecting a number
2. It receives `"arr = [-1, -2, -3, -4]"` instead
3. `int()` throws `ValueError` → runtime error

## Root Cause

In `go-backend/handlers/ai_test_builder.go`, the AI prompt instructed:
```
For two-parameter problems like Two Sum where input has both an array and a target, format the input as:
"nums = [2, 7, 11, 15]\ntarget = 9"
```

This is **Python syntax**, not stdin format. The judge pipes this directly to stdin, causing the mismatch.

## Fix Applied

### 1. Updated AI Prompt in `ai_test_builder.go`

Changed testcase format instructions to:

```
TESTCASE FORMAT RULES (CRITICAL):
Format testcase input as plain stdin data that a competitive programming solution reads via input().
NEVER use variable assignment format like: nums = [2, 7, 11, 15]
ALWAYS use plain numbers separated by spaces and newlines.

For array problems (e.g., Find Maximum Element):
Line 1: n (array size)
Line 2: space-separated elements
Example input: "4\n-1 -2 -3 -4"
Example output: "-1"

For two-parameter problems (e.g., Two Sum):
Line 1: n (array size)
Line 2: space-separated array elements
Line 3: target value
Example input: "4\n2 7 11 15\n9"
Example output: "0 1"

For string problems:
Line 1: the string
Example input: "hello"
Example output: "5"
```

### 2. Updated Example Starter Code

Changed the example in the JSON schema from:
```python
def two_sum(nums, target):
    pass
```

To proper stdin-reading code:
```python
n = int(input())
arr = list(map(int, input().split()))
# Your solution here
print()
```

### 3. Starter Code Templates Already Correct

The templates in `go-backend/judge/executor.go` `SupportedLanguages()` were already correct:

**Python:**
```python
import sys

def solve():
    n = int(input())
    arr = list(map(int, input().split()))
    # Your solution here
    print(arr)

solve()
```

**C++:**
```cpp
#include <bits/stdc++.h>
using namespace std;

int main() {
    int n;
    cin >> n;
    vector<int> arr(n);
    for(int i=0;i<n;i++) cin >> arr[i];
    // Your solution here
    return 0;
}
```

**Java:**
```java
import java.util.Scanner;

public class Main {
    public static void main(String[] args) {
        Scanner sc = new Scanner(System.in);
        int n = sc.nextInt();
        int[] arr = new int[n];
        for(int i=0;i<n;i++) arr[i] = sc.nextInt();
        // Your solution here
    }
}
```

## Manual Fix Required for Existing Tests

For the "Find Maximum Element" test that already exists in the database:

1. Go to **Admin → Tests → Find Maximum Element**
2. Edit each testcase to change format:

**Before:**
```
Input: arr = [-1, -2, -3, -4]
Expected: -1
```

**After:**
```
Input: 4
-1 -2 -3 -4
Expected: -1
```

All 3 testcases need this fix:
- Case 1: `"4\n-1 -2 -3 -4"` → expected: `"-1"`
- Case 2: `"9\n3 1 4 1 5 9 2 6 5"` → expected: `"9"`
- Case 3: `"1\n1"` → expected: `"1"`

3. Update the starter code template to show proper input reading

## Impact

✅ **Future tests** generated via AI Build Test will use correct format
✅ **Starter code templates** guide students to read input correctly
⚠️ **Existing tests** need manual correction in admin panel

## Testing

After applying this fix:
1. Upload a new PDF with coding problems via AI Build Test
2. Verify generated testcases use `"n\nelem1 elem2 ..."` format
3. Verify starter code shows proper `input()` usage
4. Submit a solution and confirm testcases pass/fail correctly

## Files Modified

- `go-backend/handlers/ai_test_builder.go` - Updated AI prompt and example format
- `go-backend/judge/executor.go` - No changes needed (already correct)
