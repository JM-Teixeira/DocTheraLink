/**
 * Sample Data Seeding Utility for Klyniva
 * 
 * This utility provides sample facilities, patients, and other data
 * to bootstrap the Klyniva healthcare management system.
 */

import { FacilityModel } from "@/components/Facility/models";
import { PatientModel } from "@/components/Patient/models";

export interface SampleFacility {
  name: string;
  facility_type: string;
  address: string;
  district: number;
  state: number;
  pincode: string;
  logo?: string;
}

export interface SamplePatient {
  name: string;
  age: number;
  gender: number;
  address: string;
  medical_history: string[];
  current_condition: string;
  facility: string;
}

export const sampleFacilities: SampleFacility[] = [
  {
    name: "Climutua Medical Center",
    facility_type: "Private Hospital",
    address: "Rua da Saúde, 123, Lisboa",
    district: 1, // This would need to be mapped to actual district IDs
    state: 1, // This would need to be mapped to actual state IDs
    pincode: "1000-001",
    logo: "/images/climutua.jpg"
  },
  {
    name: "Klyniva General Hospital",
    facility_type: "Public Hospital",
    address: "Avenida Central, 456, Porto",
    district: 2,
    state: 1,
    pincode: "4000-001"
  }
];

export const samplePatients: SamplePatient[] = [
  {
    name: "Maria Silva",
    age: 45,
    gender: 2, // Female
    address: "Rua das Flores, 789, Lisboa",
    medical_history: ["Diabetes Type 2", "Hypertension"],
    current_condition: "Routine checkup for diabetes management",
    facility: "Climutua Medical Center"
  },
  {
    name: "João Santos",
    age: 62,
    gender: 1, // Male
    address: "Praça da República, 321, Porto",
    medical_history: ["Cardiac Arrhythmia", "High Cholesterol"],
    current_condition: "Follow-up for cardiac condition",
    facility: "Climutua Medical Center"
  },
  {
    name: "Ana Costa",
    age: 28,
    gender: 2, // Female
    address: "Rua Nova, 654, Coimbra",
    medical_history: ["Asthma"],
    current_condition: "Respiratory therapy consultation",
    facility: "Klyniva General Hospital"
  },
  {
    name: "Pedro Oliveira",
    age: 55,
    gender: 1, // Male
    address: "Avenida da Liberdade, 987, Lisboa",
    medical_history: ["Arthritis", "Lower back pain"],
    current_condition: "Physical therapy for joint mobility",
    facility: "Climutua Medical Center"
  },
  {
    name: "Catarina Ferreira",
    age: 38,
    gender: 2, // Female
    address: "Rua do Comércio, 147, Braga",
    medical_history: ["Migraine", "Anxiety"],
    current_condition: "Neurological consultation",
    facility: "Klyniva General Hospital"
  }
];

export const sampleTimetableEntries = [
  {
    patient_name: "Maria Silva",
    therapist_name: "Dr. Carlos Mendes",
    session_type: "Physical Therapy",
    start_time: "09:00",
    end_time: "10:00",
    section: "Rehabilitation",
    status: "scheduled"
  },
  {
    patient_name: "João Santos",
    therapist_name: "Dr. Isabel Rodrigues",
    session_type: "Cardiac Rehabilitation",
    start_time: "10:30",
    end_time: "11:30",
    section: "Cardiology",
    status: "scheduled"
  },
  {
    patient_name: "Ana Costa",
    therapist_name: "Dr. Miguel Pereira",
    session_type: "Respiratory Therapy",
    start_time: "14:00",
    end_time: "15:00",
    section: "Pulmonology",
    status: "scheduled"
  },
  {
    patient_name: "Pedro Oliveira",
    therapist_name: "Dr. Sofia Almeida",
    session_type: "Joint Mobility",
    start_time: "15:30",
    end_time: "16:30",
    section: "Orthopedics",
    status: "scheduled"
  }
];

/**
 * Utility function to create sample data
 * Note: This is for demonstration purposes only
 * In a real application, this would interact with the backend API
 */
export const createSampleData = async () => {
  console.log("Sample data for Klyniva:");
  console.log("Facilities:", sampleFacilities);
  console.log("Patients:", samplePatients);
  console.log("Timetable Entries:", sampleTimetableEntries);
  
  // In a real implementation, you would make API calls here
  // Example:
  // for (const facility of sampleFacilities) {
  //   await request(routes.createFacility, { body: facility });
  // }
  
  return {
    facilities: sampleFacilities,
    patients: samplePatients,
    timetable: sampleTimetableEntries
  };
};