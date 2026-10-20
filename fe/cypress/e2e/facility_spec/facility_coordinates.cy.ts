/**
 * Facility Coordinates Type Fix Test Suite
 * 
 * This test suite verifies that latitude and longitude values are properly
 * converted from strings to numbers before being sent to the backend API.
 * 
 * Requirements tested:
 * - 1.1: Convert string values to numbers on form submission
 * - 1.2: Send undefined/null for empty coordinate fields
 * - 1.3: Convert coordinates on update form submission
 * - 1.4: Backend successfully parses coordinates as float64
 * - 2.1: Create facility without coordinates
 * - 2.2: Update facility to clear coordinates
 * - 2.3: Display existing coordinates properly in edit mode
 */

import LoginPage from "../../pageobject/Login/LoginPage";

describe("Facility Coordinates Type Fix", () => {
  const loginPage = new LoginPage();

  before(() => {
    loginPage.loginByRole("districtAdmin");
    cy.saveLocalStorage();
  });

  beforeEach(() => {
    cy.viewport(1280, 720);
    cy.restoreLocalStorage();
  });

  it("should create a facility with valid coordinates and send them as numbers", () => {
    // Intercept the API request to verify payload
    cy.intercept("POST", "**/api/v1/facility/").as("createFacility");

    // Navigate to facility creation page
    cy.visit("/facility/create");
    cy.wait(1000);

    // Fill in required fields
    cy.get("#facility_type").click();
    cy.get("[role='option']").contains("Primary Health Centre").click();
    
    cy.get("#name").type("Test Facility with Coordinates");
    
    cy.get("#pincode").type("1000-001");
    
    // Select state
    cy.get("#state").click();
    cy.get("[role='option']").contains("Lisboa").click();
    cy.wait(500);
    
    // Select district
    cy.get("#district").click();
    cy.get("[role='option']").first().click();
    
    cy.get("#address").type("123 Test Street");
    
    // Enter coordinates
    cy.get("#latitude").clear().type("38.7223");
    cy.get("#longitude").clear().type("-9.1393");

    // Submit the form
    cy.get("button[type='button']").contains("Save Facility").click();

    // Wait for the API call and verify the payload
    cy.wait("@createFacility").then((interception) => {
      const requestBody = interception.request.body;
      
      // Verify coordinates are sent as numbers, not strings
      expect(requestBody.latitude).to.be.a("number");
      expect(requestBody.longitude).to.be.a("number");
      expect(requestBody.latitude).to.equal(38.7223);
      expect(requestBody.longitude).to.equal(-9.1393);
      
      // Verify response is successful
      expect(interception.response?.statusCode).to.equal(201);
    });

    // Verify success notification and redirect
    cy.url().should("include", "/facility/");
    cy.verifyNotification("Facility added successfully");
  });

  it("should create a facility without coordinates (empty fields)", () => {
    cy.intercept("POST", "**/api/v1/facility/").as("createFacilityNoCoords");

    cy.visit("/facility/create");
    cy.wait(1000);

    // Fill in required fields only
    cy.get("#facility_type").click();
    cy.get("[role='option']").contains("Primary Health Centre").click();
    
    cy.get("#name").type("Test Facility without Coordinates");
    
    cy.get("#pincode").type("1000-002");
    
    cy.get("#state").click();
    cy.get("[role='option']").contains("Lisboa").click();
    cy.wait(500);
    
    cy.get("#district").click();
    cy.get("[role='option']").first().click();
    
    cy.get("#address").type("456 Test Avenue");
    
    // Leave latitude and longitude empty
    cy.get("#latitude").should("have.value", "");
    cy.get("#longitude").should("have.value", "");

    cy.get("button[type='button']").contains("Save Facility").click();

    cy.wait("@createFacilityNoCoords").then((interception) => {
      const requestBody = interception.request.body;
      
      // Verify coordinates are undefined (not sent or sent as undefined)
      expect(requestBody.latitude).to.be.undefined;
      expect(requestBody.longitude).to.be.undefined;
      
      expect(interception.response?.statusCode).to.equal(201);
    });

    cy.url().should("include", "/facility/");
    cy.verifyNotification("Facility added successfully");
  });

  it("should update a facility to add coordinates", () => {
    // First create a facility without coordinates
    cy.intercept("POST", "**/api/v1/facility/").as("createFacility");
    cy.visit("/facility/create");
    cy.wait(1000);

    cy.get("#facility_type").click();
    cy.get("[role='option']").contains("Primary Health Centre").click();
    cy.get("#name").type("Facility to Add Coords");
    cy.get("#pincode").type("1000-003");
    cy.get("#state").click();
    cy.get("[role='option']").contains("Lisboa").click();
    cy.wait(500);
    cy.get("#district").click();
    cy.get("[role='option']").first().click();
    cy.get("#address").type("789 Test Road");
    cy.get("button[type='button']").contains("Save Facility").click();

    cy.wait("@createFacility").then((interception) => {
      const facilityId = interception.response?.body.id;
      
      // Now update the facility to add coordinates
      cy.intercept("PUT", `**/api/v1/facility/${facilityId}/`).as("updateFacility");
      cy.visit(`/facility/${facilityId}/update`);
      cy.wait(1000);

      // Add coordinates
      cy.get("#latitude").clear().type("38.7500");
      cy.get("#longitude").clear().type("-9.2000");

      cy.get("button[type='button']").contains("Update Facility").click();

      cy.wait("@updateFacility").then((updateInterception) => {
        const requestBody = updateInterception.request.body;
        
        // Verify coordinates are sent as numbers
        expect(requestBody.latitude).to.be.a("number");
        expect(requestBody.longitude).to.be.a("number");
        expect(requestBody.latitude).to.equal(38.75);
        expect(requestBody.longitude).to.equal(-9.2);
        
        expect(updateInterception.response?.statusCode).to.equal(200);
      });

      cy.verifyNotification("Facility updated successfully");
    });
  });

  it("should update a facility to remove coordinates", () => {
    // First create a facility with coordinates
    cy.intercept("POST", "**/api/v1/facility/").as("createFacility");
    cy.visit("/facility/create");
    cy.wait(1000);

    cy.get("#facility_type").click();
    cy.get("[role='option']").contains("Primary Health Centre").click();
    cy.get("#name").type("Facility to Remove Coords");
    cy.get("#pincode").type("1000-004");
    cy.get("#state").click();
    cy.get("[role='option']").contains("Lisboa").click();
    cy.wait(500);
    cy.get("#district").click();
    cy.get("[role='option']").first().click();
    cy.get("#address").type("321 Test Lane");
    cy.get("#latitude").type("38.8000");
    cy.get("#longitude").type("-9.3000");
    cy.get("button[type='button']").contains("Save Facility").click();

    cy.wait("@createFacility").then((interception) => {
      const facilityId = interception.response?.body.id;
      
      // Now update the facility to remove coordinates
      cy.intercept("PUT", `**/api/v1/facility/${facilityId}/`).as("updateFacility");
      cy.visit(`/facility/${facilityId}/update`);
      cy.wait(1000);

      // Verify coordinates are loaded
      cy.get("#latitude").should("have.value", "38.8");
      cy.get("#longitude").should("have.value", "-9.3");

      // Clear coordinates
      cy.get("#latitude").clear();
      cy.get("#longitude").clear();

      cy.get("button[type='button']").contains("Update Facility").click();

      cy.wait("@updateFacility").then((updateInterception) => {
        const requestBody = updateInterception.request.body;
        
        // Verify coordinates are undefined
        expect(requestBody.latitude).to.be.undefined;
        expect(requestBody.longitude).to.be.undefined;
        
        expect(updateInterception.response?.statusCode).to.equal(200);
      });

      cy.verifyNotification("Facility updated successfully");
    });
  });

  it("should update a facility to modify existing coordinates", () => {
    // First create a facility with coordinates
    cy.intercept("POST", "**/api/v1/facility/").as("createFacility");
    cy.visit("/facility/create");
    cy.wait(1000);

    cy.get("#facility_type").click();
    cy.get("[role='option']").contains("Primary Health Centre").click();
    cy.get("#name").type("Facility to Modify Coords");
    cy.get("#pincode").type("1000-005");
    cy.get("#state").click();
    cy.get("[role='option']").contains("Lisboa").click();
    cy.wait(500);
    cy.get("#district").click();
    cy.get("[role='option']").first().click();
    cy.get("#address").type("654 Test Boulevard");
    cy.get("#latitude").type("38.6000");
    cy.get("#longitude").type("-9.1000");
    cy.get("button[type='button']").contains("Save Facility").click();

    cy.wait("@createFacility").then((interception) => {
      const facilityId = interception.response?.body.id;
      
      // Now update the facility with new coordinates
      cy.intercept("PUT", `**/api/v1/facility/${facilityId}/`).as("updateFacility");
      cy.visit(`/facility/${facilityId}/update`);
      cy.wait(1000);

      // Verify old coordinates are loaded
      cy.get("#latitude").should("have.value", "38.6");
      cy.get("#longitude").should("have.value", "-9.1");

      // Modify coordinates
      cy.get("#latitude").clear().type("38.9000");
      cy.get("#longitude").clear().type("-9.4000");

      cy.get("button[type='button']").contains("Update Facility").click();

      cy.wait("@updateFacility").then((updateInterception) => {
        const requestBody = updateInterception.request.body;
        
        // Verify new coordinates are sent as numbers
        expect(requestBody.latitude).to.be.a("number");
        expect(requestBody.longitude).to.be.a("number");
        expect(requestBody.latitude).to.equal(38.9);
        expect(requestBody.longitude).to.equal(-9.4);
        
        expect(updateInterception.response?.statusCode).to.equal(200);
      });

      cy.verifyNotification("Facility updated successfully");
    });
  });

  it("should show validation errors for invalid coordinate values", () => {
    cy.visit("/facility/create");
    cy.wait(1000);

    cy.get("#facility_type").click();
    cy.get("[role='option']").contains("Primary Health Centre").click();
    cy.get("#name").type("Test Invalid Coords");
    cy.get("#pincode").type("1000-006");
    cy.get("#state").click();
    cy.get("[role='option']").contains("Lisboa").click();
    cy.wait(500);
    cy.get("#district").click();
    cy.get("[role='option']").first().click();
    cy.get("#address").type("999 Test Street");

    // Test invalid latitude (out of range)
    cy.get("#latitude").type("999");
    cy.get("#longitude").type("-9.1393");
    cy.get("button[type='button']").contains("Save Facility").click();
    
    // Should show validation error
    cy.get("#latitude-error").should("contain", "invalid");

    // Test invalid longitude (out of range)
    cy.get("#latitude").clear().type("38.7223");
    cy.get("#longitude").clear().type("999");
    cy.get("button[type='button']").contains("Save Facility").click();
    
    cy.get("#longitude-error").should("contain", "invalid");

    // Test non-numeric values
    cy.get("#latitude").clear().type("abc");
    cy.get("#longitude").clear().type("xyz");
    cy.get("button[type='button']").contains("Save Facility").click();
    
    // Should show validation errors or prevent submission
    cy.url().should("include", "/facility/create");
  });

  it("should properly display existing coordinates in edit mode", () => {
    // Create a facility with coordinates
    cy.intercept("POST", "**/api/v1/facility/").as("createFacility");
    cy.visit("/facility/create");
    cy.wait(1000);

    cy.get("#facility_type").click();
    cy.get("[role='option']").contains("Primary Health Centre").click();
    cy.get("#name").type("Test Display Coords");
    cy.get("#pincode").type("1000-007");
    cy.get("#state").click();
    cy.get("[role='option']").contains("Lisboa").click();
    cy.wait(500);
    cy.get("#district").click();
    cy.get("[role='option']").first().click();
    cy.get("#address").type("111 Test Plaza");
    cy.get("#latitude").type("38.7223456");
    cy.get("#longitude").type("-9.1393789");
    cy.get("button[type='button']").contains("Save Facility").click();

    cy.wait("@createFacility").then((interception) => {
      const facilityId = interception.response?.body.id;
      
      // Visit the edit page
      cy.visit(`/facility/${facilityId}/update`);
      cy.wait(1000);

      // Verify coordinates are properly displayed as strings in input fields
      cy.get("#latitude").should("have.value").and("match", /38\.7223/);
      cy.get("#longitude").should("have.value").and("match", /-9\.1393/);
      
      // Verify they are displayed as strings (input fields always show strings)
      cy.get("#latitude").invoke("val").should("be.a", "string");
      cy.get("#longitude").invoke("val").should("be.a", "string");
    });
  });

  afterEach(() => {
    cy.saveLocalStorage();
  });
});
