# goboxd - Complete Testing Guide & Curl Commands

**Testing the new `/languages` and `/languages/{id}` endpoints**

---

## QUICK START - Test All Endpoints

### 1. Health Check (Always Works)
```bash
curl http://localhost:8080/healthz
```
**Expected Response:**
```json
{"status":"ok"}
```

---

## NEW ENDPOINTS - Test These First

### 2. List All Languages
```bash
curl http://localhost:8080/languages
```

**Expected Response (HTTP 200):**
```json
{
  "languages": [
    {
      "id": "py3",
      "name": "Python 3"
    },
    {
      "id": "cpp",
      "name": "C++"
    },
    {
      "id": "java",
      "name": "Java"
    },
    {
      "id": "bash",
      "name": "Bash"
    },
    {
      "id": "js",
      "name": "JavaScript"
    },
    {
      "id": "c",
      "name": "C"
    },
    {
      "id": "verilog",
      "name": "Verilog"
    },
    {
      "id": "ruby",
      "name": "Ruby"
    },
    {
      "id": "lua",
      "name": "Lua"
    },
    {
      "id": "rust",
      "name": "Rust"
    },
    {
      "id": "go",
      "name": "Go"
    },
    {
      "id": "kotlin",
      "name": "Kotlin"
    }
  ]
}
```

---

### 3. Get Language Detail - Python
```bash
curl http://localhost:8080/languages/py3
```

**Expected Response (HTTP 200):**
```json
{
  "id": "py3",
  "name": "Python 3",
  "compiled": false,
  "extension": ".py"
}
```

---

### 4. Get Language Detail - C++
```bash
curl http://localhost:8080/languages/cpp
```

**Expected Response (HTTP 200):**
```json
{
  "id": "cpp",
  "name": "C++",
  "compiled": true,
  "extension": ".cpp"
}
```

---

### 5. Get Language Detail - Unknown Language
```bash
curl http://localhost:8080/languages/unknown
```

**Expected Response (HTTP 404):**
```json
{
  "error": {
    "code": "not_found",
    "message": "language not found: unknown"
  }
}
```

---

### 6. Try POST on GET-only Endpoint
```bash
curl -X POST http://localhost:8080/languages \
  -H "Content-Type: application/json" \
  -d '{}'
```

**Expected Response (HTTP 405):**
```json
{
  "error": {
    "code": "method_not_allowed",
    "message": "method not allowed"
  }
}
```

---

## EXISTING ENDPOINTS - Verify No Regressions

### 7. Readiness Probe
```bash
curl http://localhost:8080/readyz
```

**Expected Response (HTTP 200 or 503):**
```json
{
  "status": "ok",
  "languages": {
    "py3": {
      "ok": true,
      "version": "Python 3.X.X"
    },
    "cpp": {
      "ok": true,
      "version": "..."
    }
  }
}
```

---

### 8. Server Info
```bash
curl http://localhost:8080/info
```

**Expected Response (HTTP 200):**
```json
{
  "build_info": {
    "version": "0.1.0",
    "go_version": "go1.22.3"
  },
  "languages": [
    {
      "id": "py3",
      "name": "Python 3"
    },
    ...
  ],
  "limits": {
    "max_concurrent_jobs": 8,
    "max_source_bytes": 262144,
    "max_tests": 50
  }
}
```

---

## CODE EXECUTION - Test /run Endpoint

### 9. Execute Python Code - Success
```bash
curl -X POST http://localhost:8080/run \
  -H "Content-Type: application/json" \
  -d '{
    "language": "py3",
    "source": "print(\"hello world\")",
    "tests": [
      {
        "stdin": "",
        "expected_stdout": "hello world\n"
      }
    ]
  }'
```

**Expected Response (HTTP 200):**
```json
{
  "status": "accepted",
  "tests": [
    {
      "status": "accepted",
      "stdout": "hello world\n",
      "stderr": "",
      "duration_ms": 45
    }
  ]
}
```

---

### 10. Execute Python Code - Wrong Output
```bash
curl -X POST http://localhost:8080/run \
  -H "Content-Type: application/json" \
  -d '{
    "language": "py3",
    "source": "print(\"goodbye\")",
    "tests": [
      {
        "stdin": "",
        "expected_stdout": "hello world\n"
      }
    ]
  }'
```

**Expected Response (HTTP 200):**
```json
{
  "status": "wrong_output",
  "tests": [
    {
      "status": "wrong_output",
      "stdout": "goodbye\n",
      "stderr": "",
      "duration_ms": 40
    }
  ]
}
```

---

### 11. Compile & Execute C++
```bash
curl -X POST http://localhost:8080/run \
  -H "Content-Type: application/json" \
  -d '{
    "language": "cpp",
    "source": "#include <iostream>\nint main() { std::cout << \"test\"; return 0; }",
    "source_filename": "main.cpp",
    "artifact_filename": "solution",
    "build": {
      "limits": {
        "wall_time_s": 30,
        "memory_kb": 1048576,
        "max_processes": 100
      }
    },
    "run": {
      "limits": {
        "wall_time_s": 10,
        "memory_kb": 524288,
        "max_processes": 64
      }
    },
    "tests": [
      {
        "stdin": "",
        "expected_stdout": "test"
      }
    ]
  }'
```

**Expected Response (HTTP 200):**
```json
{
  "status": "accepted",
  "build": {
    "status": "ok",
    "stdout": "",
    "stderr": "",
    "duration_ms": 250
  },
  "tests": [
    {
      "status": "accepted",
      "stdout": "test",
      "stderr": "",
      "duration_ms": 15
    }
  ]
}
```

---

### 12. Execute Go Code
```bash
curl -X POST http://localhost:8080/run \
  -H "Content-Type: application/json" \
  -d '{
    "language": "go",
    "source": "package main\nimport \"fmt\"\nfunc main() { fmt.Println(\"hello\") }",
    "source_filename": "main.go",
    "artifact_filename": "solution",
    "build": {
      "limits": {
        "wall_time_s": 30,
        "memory_kb": 1048576,
        "max_processes": 100
      }
    },
    "run": {
      "limits": {
        "wall_time_s": 10,
        "memory_kb": 524288,
        "max_processes": 64
      }
    },
    "tests": [
      {
        "stdin": "",
        "expected_stdout": "hello\n"
      }
    ]
  }'
```

---

### 13. Compile Error Test
```bash
curl -X POST http://localhost:8080/run \
  -H "Content-Type: application/json" \
  -d '{
    "language": "cpp",
    "source": "#include <iostream>\nint main() { invalid syntax }",
    "source_filename": "main.cpp",
    "artifact_filename": "solution",
    "build": {
      "limits": {
        "wall_time_s": 30,
        "memory_kb": 1048576,
        "max_processes": 100
      }
    },
    "run": {
      "limits": {
        "wall_time_s": 10,
        "memory_kb": 524288,
        "max_processes": 64
      }
    },
    "tests": [
      {
        "stdin": "",
        "expected_stdout": ""
      }
    ]
  }'
```

**Expected Response (HTTP 200):**
```json
{
  "status": "build_failed",
  "build": {
    "status": "failed",
    "stdout": "",
    "stderr": "... compilation error details ...",
    "duration_ms": 180
  },
  "tests": [
    {
      "status": "not_executed",
      "stdout": "",
      "stderr": "",
      "duration_ms": 0
    }
  ]
}
```

---

### 14. Unknown Language Error
```bash
curl -X POST http://localhost:8080/run \
  -H "Content-Type: application/json" \
  -d '{
    "language": "xyz",
    "source": "test",
    "tests": [{"stdin": "", "expected_stdout": ""}]
  }'
```

**Expected Response (HTTP 400):**
```json
{
  "error": {
    "code": "unknown_language",
    "message": "language not supported: xyz"
  }
}
```

---

### 15. Missing Language Error
```bash
curl -X POST http://localhost:8080/run \
  -H "Content-Type: application/json" \
  -d '{
    "source": "test",
    "tests": [{"stdin": "", "expected_stdout": ""}]
  }'
```

**Expected Response (HTTP 400):**
```json
{
  "error": {
    "code": "missing_language",
    "message": "language is required"
  }
}
```

---

### 16. Missing Tests Error
```bash
curl -X POST http://localhost:8080/run \
  -H "Content-Type: application/json" \
  -d '{
    "language": "py3",
    "source": "print(1)",
    "tests": []
  }'
```

**Expected Response (HTTP 400):**
```json
{
  "error": {
    "code": "missing_tests",
    "message": "at least one test is required"
  }
}
```

---

### 17. Path Traversal Protection Test
```bash
curl -X POST http://localhost:8080/run \
  -H "Content-Type: application/json" \
  -d '{
    "language": "py3",
    "source": "print(1)",
    "source_filename": "../../etc/passwd",
    "tests": [{"stdin": "", "expected_stdout": ""}]
  }'
```

**Expected Response (HTTP 400):**
```json
{
  "error": {
    "code": "invalid_filename",
    "message": "source_filename must be a single path component"
  }
}
```

---

## AUTOMATION - Test Script

Run all tests automatically:
```bash
bash test_endpoints.sh http://localhost:8080
```

Or using Python:
```python
import requests
import json

BASE = "http://localhost:8080"

# Test 1: List languages
resp = requests.get(f"{BASE}/languages")
assert resp.status_code == 200
languages = resp.json()
print(f"Found {len(languages['languages'])} languages")

# Test 2: Get language detail
resp = requests.get(f"{BASE}/languages/py3")
assert resp.status_code == 200
lang = resp.json()
assert lang['id'] == 'py3'
assert lang['compiled'] == False

# Test 3: Unknown language
resp = requests.get(f"{BASE}/languages/xyz")
assert resp.status_code == 404

# Test 4: Execute code
payload = {
    "language": "py3",
    "source": "print('test')",
    "tests": [{"stdin": "", "expected_stdout": "test\n"}]
}
resp = requests.post(f"{BASE}/run", json=payload)
assert resp.status_code == 200
result = resp.json()
assert result['status'] == 'accepted'

print("All tests passed!")
```

---

## PERFORMANCE TESTING

### Concurrent Requests
```bash
# Send 10 concurrent Python jobs
for i in {1..10}; do
  curl -X POST http://localhost:8080/run \
    -H "Content-Type: application/json" \
    -d '{"language":"py3","source":"import time; time.sleep(0.1)","tests":[{"stdin":"","expected_stdout":""}]}' &
done
wait
echo "All concurrent jobs completed"
```

---

## TROUBLESHOOTING

### 404 on /languages endpoint
**Issue:** Old server binary running without endpoint registration  
**Solution:** Rebuild binary and restart server
```bash
go build -o goboxd ./cmd/goboxd
./goboxd
```

### Connection refused
**Issue:** Server not listening on port 8080  
**Solution:** Check if server is running
```bash
ps aux | grep goboxd
# If not running, start with: ./goboxd or docker run -p 8080:8080 goboxd:latest
```

### Timeout on /languages/{id}
**Issue:** Server processing is slow  
**Solution:** Not a timeout (Go HTTP has no default timeout)  
**Normal:** /languages endpoints should respond in <10ms

### Request body too large
**Issue:** Posted JSON exceeds 256 KiB limit  
**Solution:** Reduce source code size or number of test cases

---

## DOCKER TESTING

If WSL testing is problematic, use Docker:

```bash
# Build image
docker build -t goboxd:latest .

# Run container
docker run -p 8080:8080 goboxd:latest

# In another terminal, test
curl http://localhost:8080/languages
```

The Docker approach automatically handles:
- Correct process ownership
- Network isolation
- nsjail availability
- System libraries and compilers
- Reproducible environment

