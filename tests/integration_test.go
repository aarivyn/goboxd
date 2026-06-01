package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

const baseURL = "http://localhost:8080"

type IntTestCase struct {
	Stdin          string `json:"stdin"`
	ExpectedStdout string `json:"expected_stdout"`
}

type IntRunReq struct {
	Language         string        `json:"language"`
	Source           string        `json:"source"`
	SourceFilename   string        `json:"source_filename,omitempty"`
	ArtifactFilename string        `json:"artifact_filename,omitempty"`
	Tests            []IntTestCase `json:"tests"`
}

type IntRunResp struct {
	Status string `json:"status"`
}

func runCode(t *testing.T, req IntRunReq) IntRunResp {
	t.Helper()
	body, _ := json.Marshal(req)
	resp, err := http.Post(baseURL+"/run", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	var result IntRunResp
	json.NewDecoder(resp.Body).Decode(&result)
	return result
}

func TestIntHealthz(t *testing.T) {
	resp, err := http.Get(baseURL + "/healthz")
	if err != nil {
		t.Fatalf("healthz failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}
}

func TestIntPython(t *testing.T) {
	resp := runCode(t, IntRunReq{
		Language: "py3",
		Source:   "print('hello')",
		Tests:    []IntTestCase{{Stdin: "", ExpectedStdout: "hello"}},
	})
	if resp.Status != "accepted" {
		t.Errorf("expected accepted got %s", resp.Status)
	}
}

func TestIntCpp(t *testing.T) {
	resp := runCode(t, IntRunReq{
		Language: "cpp",
		Source:   "#include<iostream>\nint main(){std::cout<<\"hello\";return 0;}",
		Tests:    []IntTestCase{{Stdin: "", ExpectedStdout: "hello"}},
	})
	if resp.Status != "accepted" {
		t.Errorf("expected accepted got %s", resp.Status)
	}
}

func TestIntC(t *testing.T) {
	resp := runCode(t, IntRunReq{
		Language: "c",
		Source:   "#include<stdio.h>\nint main(){printf(\"hello\");return 0;}",
		Tests:    []IntTestCase{{Stdin: "", ExpectedStdout: "hello"}},
	})
	if resp.Status != "accepted" {
		t.Errorf("expected accepted got %s", resp.Status)
	}
}

func TestIntJava(t *testing.T) {
	resp := runCode(t, IntRunReq{
		Language:         "java",
		Source:           "public class Main{public static void main(String[] args){System.out.print(\"hello\");}}",
		SourceFilename:   "Main.java",
		ArtifactFilename: "Main",
		Tests:            []IntTestCase{{Stdin: "", ExpectedStdout: "hello"}},
	})
	if resp.Status != "accepted" {
		t.Errorf("expected accepted got %s", resp.Status)
	}
}

func TestIntBash(t *testing.T) {
	resp := runCode(t, IntRunReq{
		Language: "bash",
		Source:   "echo hello",
		Tests:    []IntTestCase{{Stdin: "", ExpectedStdout: "hello"}},
	})
	if resp.Status != "accepted" {
		t.Errorf("expected accepted got %s", resp.Status)
	}
}

func TestIntJS(t *testing.T) {
	resp := runCode(t, IntRunReq{
		Language: "js",
		Source:   "console.log('hello')",
		Tests:    []IntTestCase{{Stdin: "", ExpectedStdout: "hello"}},
	})
	if resp.Status != "accepted" {
		t.Errorf("expected accepted got %s", resp.Status)
	}
}

func TestIntWrongOutput(t *testing.T) {
	resp := runCode(t, IntRunReq{
		Language: "py3",
		Source:   "print('wrong')",
		Tests:    []IntTestCase{{Stdin: "", ExpectedStdout: "hello"}},
	})
	if resp.Status != "wrong_output" {
		t.Errorf("expected wrong_output got %s", resp.Status)
	}
}

func TestIntUnknownLanguage(t *testing.T) {
	body, _ := json.Marshal(IntRunReq{
		Language: "nonexistent",
		Source:   "print('hi')",
		Tests:    []IntTestCase{{Stdin: "", ExpectedStdout: "hi"}},
	})
	resp, _ := http.Post(baseURL+"/run", "application/json", bytes.NewReader(body))
	if resp.StatusCode != 400 {
		t.Errorf("expected 400 got %d", resp.StatusCode)
	}
}

func TestIntInvalidFilename(t *testing.T) {
	body, _ := json.Marshal(map[string]interface{}{
		"language":        "cpp",
		"source":          "#include<iostream>\nint main(){}",
		"source_filename": "../../etc/passwd",
		"tests":           []IntTestCase{{Stdin: "", ExpectedStdout: ""}},
	})
	resp, _ := http.Post(baseURL+"/run", "application/json", bytes.NewReader(body))
	if resp.StatusCode != 400 {
		t.Errorf("expected 400 got %d", resp.StatusCode)
	}
}
func TestIntDisallowedFlag(t *testing.T) {
	body, _ := json.Marshal(map[string]interface{}{
		"language": "cpp",
		"source": "#include<iostream>\nint main(){std::cout<<\"hello\";return 0;}",
		"build": map[string]interface{}{
			"flags": []string{"-nostdlib"},
		},
		"tests": []IntTestCase{{Stdin: "", ExpectedStdout: "hello"}},
	})
	resp, _ := http.Post(baseURL+"/run", "application/json", bytes.NewReader(body))
	if resp.StatusCode != 400 {
		t.Errorf("expected 400 for disallowed flag, got %d", resp.StatusCode)
	}
}

func TestIntRequestSizeLimit(t *testing.T) {
	hugeSource := string(make([]byte, 300*1024))
	body, _ := json.Marshal(IntRunReq{
		Language: "py3",
		Source:   hugeSource,
		Tests:    []IntTestCase{{Stdin: "", ExpectedStdout: ""}},
	})
	
	resp, _ := http.Post(baseURL+"/run", "application/json", bytes.NewReader(body))
	if resp.StatusCode != 413 {
		t.Errorf("expected 413 for oversized request, got %d", resp.StatusCode)
	}
}

func TestIntOutputTruncation(t *testing.T) {
	source := `print("x" * (10 * 1024 * 1024))`
	
	body, _ := json.Marshal(IntRunReq{
		Language: "py3",
		Source:   source,
		Tests:    []IntTestCase{{Stdin: "", ExpectedStdout: "x"}},
	})
	
	resp, err := http.Post(baseURL+"/run", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	
	var result struct {
		Status string `json:"status"`
		Tests  []struct {
			Stdout string `json:"stdout"`
		} `json:"tests"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	
	if len(result.Tests) > 0 {
		stdout := result.Tests[0].Stdout
		if len(stdout) > 4*1024*1024 {
			t.Errorf("expected truncated output (max 4 MiB), got %d bytes", len(stdout))
		}
		if !bytes.Contains([]byte(stdout), []byte("[output truncated]")) {
			t.Error("expected truncation marker in output")
		}
	}
}