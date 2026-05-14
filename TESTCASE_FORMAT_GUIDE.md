# Testcase Format Guide for SkillSprint OJT

## Standard Competitive Programming Input Format

All testcases must use **plain stdin format** (numbers and text only), NOT Python/Java variable syntax.

---

## ✅ Correct Formats

### 1. Single Array Problem
**Example:** Find Maximum Element

**Input Format:**
```
Line 1: n (array size)
Line 2: space-separated elements
```

**Testcase:**
```
4
-1 -2 -3 -4
```

**Expected Output:**
```
-1
```

**Python Code:**
```python
n = int(input())
arr = list(map(int, input().split()))
print(max(arr))
```

---

### 2. Array + Target Problem
**Example:** Two Sum

**Input Format:**
```
Line 1: n (array size)
Line 2: space-separated array elements
Line 3: target value
```

**Testcase:**
```
4
2 7 11 15
9
```

**Expected Output:**
```
0 1
```

**Python Code:**
```python
n = int(input())
arr = list(map(int, input().split()))
target = int(input())
# Your solution here
print(0, 1)
```

---

### 3. String Problem
**Example:** String Length

**Input Format:**
```
Line 1: the string
```

**Testcase:**
```
hello
```

**Expected Output:**
```
5
```

**Python Code:**
```python
s = input()
print(len(s))
```

---

### 4. Multiple Test Cases
**Example:** Sum of Two Numbers (T test cases)

**Input Format:**
```
Line 1: T (number of test cases)
Next T lines: a b (two integers)
```

**Testcase:**
```
3
1 2
5 7
10 20
```

**Expected Output:**
```
3
12
30
```

**Python Code:**
```python
T = int(input())
for _ in range(T):
    a, b = map(int, input().split())
    print(a + b)
```

---

## ❌ Wrong Formats (DO NOT USE)

### Variable Assignment Format
```
arr = [-1, -2, -3, -4]
```
❌ This is Python syntax, not stdin!

### JSON Format
```
{"arr": [-1, -2, -3, -4]}
```
❌ This is JSON, not stdin!

### Comma-Separated Without Spaces
```
-1,-2,-3,-4
```
❌ Use spaces, not commas!

---

## Language-Specific Input Reading

### Python
```python
n = int(input())
arr = list(map(int, input().split()))
```

### C++
```cpp
int n;
cin >> n;
vector<int> arr(n);
for(int i=0; i<n; i++) cin >> arr[i];
```

### Java
```java
Scanner sc = new Scanner(System.in);
int n = sc.nextInt();
int[] arr = new int[n];
for(int i=0; i<n; i++) arr[i] = sc.nextInt();
```

### JavaScript (Node.js)
```javascript
const lines = require('fs').readFileSync('/dev/stdin','utf8').trim().split('\n');
const n = parseInt(lines[0]);
const arr = lines[1].split(' ').map(Number);
```

### Go
```go
reader := bufio.NewReader(os.Stdin)
var n int
fmt.Fscan(reader, &n)
arr := make([]int, n)
for i := 0; i < n; i++ {
    fmt.Fscan(reader, &arr[i])
}
```

---

## Quick Checklist

When creating testcases, verify:

- [ ] Input uses only numbers, spaces, and newlines
- [ ] No variable names (arr, nums, target, etc.)
- [ ] No brackets `[]` or braces `{}`
- [ ] No equals signs `=`
- [ ] First line is array size `n` for array problems
- [ ] Elements are space-separated on second line
- [ ] Expected output matches what `print()` would produce

---

## How to Fix Existing Tests

1. Go to **Admin Panel → Tests**
2. Select the test with wrong format
3. Click on each coding question
4. Edit each testcase:
   - Change input from `arr = [1, 2, 3]` to `3\n1 2 3`
   - Verify expected output is plain text (no brackets)
5. Update starter code to show proper input reading
6. Save changes

---

## AI Build Test Behavior

After the fix, AI Build Test will automatically:
- Generate testcases in correct stdin format
- Use `n\nelem1 elem2 ...` for arrays
- Provide starter code with proper `input()` calls
- Never use variable assignment syntax

No manual conversion needed for new tests!
