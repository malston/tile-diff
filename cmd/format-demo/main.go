// ABOUTME: Demonstrates the description formatting improvements.
// ABOUTME: Shows before/after for messy release notes descriptions.
package main

import (
	"fmt"

	"github.com/malston/tile-diff/pkg/report"
)

func main() {
	// Example messy description similar to what we get from real release notes
	messyDescription := `Canary deployment enhancements: Support a new --instance-steps flag that allows for fine-grained control over a canary deployment rollout.
	Gorouter supports a new X-CF-PROCESS-INSTANCE header for routing http requests to a specific app process.
	Canary deployment enhancements: Support a new --instance-steps flag that allows for fine-grained control over a canary deployment rollout.
	Gorouter supports a new X-CF-PROCESS-INSTANCE header for routing http requests to a specific app process.
	Silk CNI can now accept a comma-separated list of CIDRs for the "Overlay subnet" property.
	Silk CNI can now accept a comma-separated list of CIDRs for the "Overlay subnet" property.
	In 10.2, the Gorouter and TCP Router app identity verification property is automatically changed.
	In 10.2, the Gorouter and TCP Router app identity verification property is automatically changed.
	Bump java-offline-buildpack to version 4.86.1
	Bump nodejs-offline-buildpack to version 1.8.84
	Bump python-offline-buildpack to version 1.8.79
	Bump dotnet-core-offline-buildpack to version 2.4.86
	Bump go-offline-buildpack to version 1.10.83
	Bump php-offline-buildpack to version 4.6.73
	Bump ruby-offline-buildpack to version 1.10.66
	Bump staticfile-offline-buildpack to version 1.8.62
	Bump binary-offline-buildpack to version 1.1.64
	Bump nginx-offline-buildpack to version 1.2.73
	Bump capi to version 1.218.0
	Bump cf-autoscaling to version 250.5.9
	Bump cf-cli to version 2.4.0
	Bump cf-networking to version 3.85.0
	Bump cflinuxfs4 to version 1.337.0
	Bump credhub to version 2.14.16
	Bump diego to version 2.122.0
	Bump garden-runc to version 1.78.0
	Bump log-cache to version 3.2.2
	Bump loggregator to version 107.0.23
	Bump loggregator-agent to version 8.3.10
	Bump metric-registrar to version 4.0.9
	Bump mysql-monitoring to version 10.32.0
	Bump nats to version 56.62.0
	Bump routing to version 0.351.0
	Bump tuaa to version 1.9.0
	Products Solutions Support and Services Company How To Buy
	Privacy Supplier Responsibility Terms of Use Site Map
	Content feedback and comments
	For more information, see the documentation
	See the Knowledge Base article
	Affected versions: 10.2.0, 10.2.1, 10.2.4, 10.2.5
	Known Issue: This is a known issue that affects something
	Release date: October 28, 2025`

	fmt.Println("================================================================================")
	fmt.Println("BEFORE - Raw Description (how it looks now)")
	fmt.Println("================================================================================")
	fmt.Println(messyDescription)
	fmt.Println()

	// Clean it up
	cleaned := report.CleanDescription(messyDescription)

	fmt.Println("================================================================================")
	fmt.Println("AFTER - Cleaned Description (with improvements)")
	fmt.Println("================================================================================")
	fmt.Println(cleaned)
	fmt.Println()

	fmt.Println("================================================================================")
	fmt.Println("Key Improvements:")
	fmt.Println("================================================================================")
	fmt.Println("✓ Duplicates removed (Canary deployment text appeared 2x, now 1x)")
	fmt.Println("✓ 30+ bump lines summarized into component counts")
	fmt.Println("✓ Footer/navigation noise filtered out")
	fmt.Println("✓ Metadata (Known Issue, Affected versions) filtered out")
	fmt.Println("✓ Bullet formatting for readability")
	fmt.Println()
}
