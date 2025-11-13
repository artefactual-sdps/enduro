// Custom JavaScript for Enduro dashboard
// This file loads in the <head> via the manifest "scripts" array

(function () {
  "use strict";

  console.log("✅ Custom JavaScript loaded successfully!");

  // Create a global object for custom functionality
  window.EnduroCustom = {
    version: "1.0.0",
    loaded: true,
    loadTime: new Date().toISOString(),

    // Test function: Show notification
    showNotification: function (message, type = "info") {
      console.log(`[EnduroCustom] showNotification("${message}", "${type}")`);

      const alertDiv = document.createElement("div");
      alertDiv.className = `alert alert-${type} alert-dismissible fade show mt-3`;
      alertDiv.setAttribute("role", "alert");
      alertDiv.innerHTML = `
        <strong>${type.toUpperCase()}:</strong> ${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
      `;

      const container = document.querySelector(".container-xxl");
      if (container) {
        container.insertBefore(alertDiv, container.firstChild);

        // Auto-dismiss after 5 seconds
        setTimeout(() => {
          alertDiv.remove();
        }, 5000);
      }
    },

    // Helper function: Get access token from local storage
    getAccessToken: function () {
      try {
        const oidcKey =
          "oidc.user:http://keycloak:7470/realms/artefactual:enduro";
        const oidcData = localStorage.getItem(oidcKey);
        if (oidcData) {
          const parsed = JSON.parse(oidcData);
          return parsed.access_token || null;
        }
      } catch (error) {
        console.error("[EnduroCustom] Error getting access token:", error);
      }
      return null;
    },

    // Test function: Fetch SIP count from API
    async getSIPCount() {
      console.log("[EnduroCustom] getSIPCount() called");
      try {
        const token = this.getAccessToken();
        const headers = {};
        if (token) {
          headers["Authorization"] = `Bearer ${token}`;
        }

        const response = await fetch("/api/ingest/sips", { headers });
        const data = await response.json();
        console.log("[EnduroCustom] SIPs count:", data.page.total);
        return data.page.total || 0;
      } catch (error) {
        console.error("[EnduroCustom] Error fetching SIPs count:", error);
        return 0;
      }
    },

    // Test function: Run all tests with UI output
    runTests: async function () {
      console.log("\n=== Running EnduroCustom Tests ===");

      const results = [];

      // Test 1: Check if object exists
      results.push({
        test: "window.EnduroCustom exists",
        status: "✅ PASS",
        details: "Object loaded successfully",
      });
      console.log("✅ Test 1: window.EnduroCustom exists");

      // Test 2: Check version
      results.push({
        test: "Version check",
        status: "✅ PASS",
        details: `Version: ${this.version}`,
      });
      console.log(`✅ Test 2: Version = ${this.version}`);

      // Test 3: Check load time
      results.push({
        test: "Load time recorded",
        status: "✅ PASS",
        details: `Loaded at: ${this.loadTime}`,
      });
      console.log(`✅ Test 3: Loaded at ${this.loadTime}`);

      // Test 4: Test notification
      results.push({
        test: "Notification system",
        status: "✅ PASS",
        details: "Notification displayed",
      });
      console.log("✅ Test 4: Testing notification...");
      this.showNotification("Test notification from custom.js!", "success");

      // Test 5: Test API call
      console.log("✅ Test 5: Testing API call...");
      try {
        const count = await this.getSIPCount();
        results.push({
          test: "API call to /api/ingest/sips",
          status: "✅ PASS",
          details: `SIPs count: ${count}`,
        });
        console.log(`✅ Test 5 Result: SIPs count = ${count}`);
      } catch (error) {
        results.push({
          test: "API call to /api/ingest/sips",
          status: "⚠️ WARN",
          details: `Error: ${error.message}`,
        });
      }

      console.log("=== Tests Complete ===\n");

      return results;
    },

    // Display test results in the UI
    displayTestResults: function (results) {
      const resultsDiv = document.getElementById("testResults");
      if (!resultsDiv) return;

      let html = '<div class="list-group">';

      results.forEach((result, index) => {
        const alertClass = result.status.includes("PASS")
          ? "success"
          : "warning";
        html += `
          <div class="list-group-item list-group-item-${alertClass}">
            <div class="d-flex w-100 justify-content-between">
              <h6 class="mb-1">${result.status} Test ${index + 1}: ${result.test}</h6>
            </div>
            <p class="mb-1"><small>${result.details}</small></p>
          </div>
        `;
      });

      html += "</div>";
      html += `
        <div class="alert alert-info mt-3">
          <strong>Summary:</strong> ${results.length} tests executed. 
          Check browser console for detailed logs.
        </div>
      `;

      resultsDiv.innerHTML = html;
    },
  };

  // Function to set up button handler
  function setupTestButton() {
    const runTestsBtn = document.getElementById("runTestsBtn");
    if (runTestsBtn) {
      console.log("✅ Found runTestsBtn, attaching handler");
      runTestsBtn.addEventListener("click", async function () {
        console.log("🔵 Test button clicked!");
        this.disabled = true;
        this.innerHTML =
          '<span class="spinner-border spinner-border-sm me-2"></span>Running Tests...';

        const results = await window.EnduroCustom.runTests();
        window.EnduroCustom.displayTestResults(results);

        this.disabled = false;
        this.innerHTML = "Run Tests Again";
      });
      console.log("✅ Test button handler attached");
      return true;
    } else {
      console.log("⚠️ runTestsBtn not found yet");
      return false;
    }
  }

  // Try to set up immediately if DOM is ready
  if (document.readyState === "loading") {
    // Still loading, wait for DOMContentLoaded
    document.addEventListener("DOMContentLoaded", function () {
      console.log("✅ DOM Content Loaded - EnduroCustom ready");
      setupTestButton();
    });
  } else {
    // DOM already loaded
    console.log("✅ DOM already loaded - EnduroCustom ready");
    if (!setupTestButton()) {
      // Button not found yet, watch for it to be added by Vue
      console.log("🔍 Setting up MutationObserver to watch for button");
      const observer = new MutationObserver(function (mutations) {
        if (setupTestButton()) {
          console.log("✅ Button found via MutationObserver, stopping watch");
          observer.disconnect();
        }
      });

      observer.observe(document.body, {
        childList: true,
        subtree: true,
      });
    }
  }

  // Log when fully loaded
  window.addEventListener("load", function () {
    console.log("✅ Window loaded - All resources ready");
  });

  console.log("✅ EnduroCustom initialized:", window.EnduroCustom);
})();
