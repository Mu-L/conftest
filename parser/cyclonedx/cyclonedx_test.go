package cyclonedx

import (
	"testing"

	"github.com/pkg/errors"
)

func TestCycloneDXParserValid(t *testing.T) {
	// generated by syft packages docker:devopps/busybox:latest -o cyclonedx --file report.xml
	// cat report.xml | cyclonedx-cli convert --input-format xml --output-format json
	sbom := `{
  "bomFormat": "CycloneDX",
  "specVersion": "1.3",
  "serialNumber": "urn:uuid:caa542d6-b9bf-469b-b564-c9ddc1e6945f",
  "version": 1,
  "metadata": {
    "timestamp": "2021-11-26T07:32:14Z",
    "tools": [
      {
        "vendor": "anchore",
        "name": "syft",
        "version": "0.30.1",
        "hashes": []
      }
    ],
    "component": {
      "type": "container",
      "name": "devopps/busybox:latest",
      "version": "sha256:d7ec60cf8390612b360c857688b383068b580d9a6ab78417c9493170ad3f1616",
      "hashes": [],
      "licenses": [],
      "externalReferences": [],
      "components": []
    }
  },
  "components": []
}
`

	parser := &Parser{}

	var input any
	if err := parser.Unmarshal([]byte(sbom), &input); err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Error("There should be information parsed but its nil")
	}

	//#nosec until https://github.com/securego/gosec/issues/1001 is fixed
	expectedSHA256 := "sha256:d7ec60cf8390612b360c857688b383068b580d9a6ab78417c9493170ad3f1616"

	metadata := input.(map[string]any)["metadata"]
	component := metadata.(map[string]any)["component"]
	currentSHA256 := component.(map[string]any)["version"]

	if expectedSHA256 != currentSHA256 {
		t.Fatalf("current SHA256 %s is different from the expected SHA256 %s", currentSHA256, expectedSHA256)
	}
}

func TestCycloneDXParserInValid(t *testing.T) {
	// generated by syft packages docker:devopps/busybox:latest -o cyclonedx --file report.xml
	// cat report.xml | cyclonedx-cli convert --input-format xml --output-format json
	sbom := `{
  "bomFormat": "CycloneDX",
  "specVersion": "1.3",
  "serialNumber": "urn:uuid:caa542d6-b9bf-469b-b564-c9ddc1e6945f",
  "version": 1,
  "metadata": {
    "timestamp": "2021-11-26T07:32:14Z",
    "tools": [
      {
        "vendor": "anchore",
        "name": "syft",
        "version": "0.30.1",
        "hashes": []
      }
    ],
    "component": {
      "type": "container",
      "name": "devopps/busybox:latest",
      "version": "COMPROMISED",
      "hashes": [],
      "licenses": [],
      "externalReferences": [],
      "components": []
    }
  },
  "components": []
}
`

	parser := &Parser{}

	var input any
	if err := parser.Unmarshal([]byte(sbom), &input); err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Error("There should be information parsed but its nil")
	}

	//#nosec until https://github.com/securego/gosec/issues/1001 is fixed
	expectedSHA256 := "sha256:d7ec60cf8390612b360c857688b383068b580d9a6ab78417c9493170ad3f1616"

	metadata := input.(map[string]any)["metadata"]
	component := metadata.(map[string]any)["component"]
	currentSHA256 := component.(map[string]any)["version"]

	var err error
	if expectedSHA256 != currentSHA256 {
		err = errors.Errorf("current SHA256 %s is different from the expected SHA256 %s", currentSHA256, expectedSHA256)
	}

	if err == nil {
		t.Error("current SHA256 and expected SHA256 should not be equal")
	}
}
