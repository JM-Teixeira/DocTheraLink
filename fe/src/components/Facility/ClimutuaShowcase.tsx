/**
 * ClimutuaShowcase Component
 * 
 * Showcases the Climutua Medical Center facility with sample data
 * Demonstrates the Klyniva healthcare management system capabilities
 */

import React from 'react';
import { useTranslation } from 'react-i18next';
import { MapPin, Phone, Users, Calendar, Activity } from 'lucide-react';

import CareIcon from "@/CAREUI/icons/CareIcon";
import { samplePatients, sampleTimetableEntries } from "@/Utils/sampleData";

const ClimutuaShowcase: React.FC = () => {
  const { t } = useTranslation();

  const facilityInfo = {
    name: "Climutua Medical Center",
    address: "Rua da Saúde, 123, Lisboa",
    phone: "+351 21 123 4567",
    type: "Private Hospital",
    logo: "/images/climutua.jpg"
  };

  const climutuaPatients = samplePatients.filter(
    patient => patient.facility === "Climutua Medical Center"
  );

  const todaysSessions = sampleTimetableEntries.slice(0, 3);

  return (
    <div className="bg-white rounded-lg shadow-lg border border-gray-200 overflow-hidden">
      {/* Header with Logo */}
      <div className="bg-gradient-to-r from-blue-600 to-blue-800 p-6 text-white">
        <div className="flex items-center space-x-4">
          <div className="w-16 h-16 bg-white rounded-lg p-2 flex items-center justify-center">
            <img
              src={facilityInfo.logo}
              alt="Climutua Logo"
              className="w-full h-full object-contain rounded"
            />
          </div>
          <div>
            <h1 className="text-2xl font-bold">{facilityInfo.name}</h1>
            <p className="text-blue-100">{facilityInfo.type}</p>
          </div>
        </div>
      </div>

      {/* Facility Information */}
      <div className="p-6 border-b border-gray-200">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">
          {t('facility_information') || 'Facility Information'}
        </h2>
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="flex items-center space-x-3">
            <MapPin className="h-5 w-5 text-gray-500" />
            <span className="text-gray-700">{facilityInfo.address}</span>
          </div>
          
          <div className="flex items-center space-x-3">
            <Phone className="h-5 w-5 text-gray-500" />
            <span className="text-gray-700">{facilityInfo.phone}</span>
          </div>
        </div>
      </div>

      {/* Statistics */}
      <div className="p-6 border-b border-gray-200">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">
          {t('facility_statistics') || 'Facility Statistics'}
        </h2>
        
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="bg-blue-50 p-4 rounded-lg text-center">
            <Users className="h-8 w-8 text-blue-600 mx-auto mb-2" />
            <div className="text-2xl font-bold text-blue-600">{climutuaPatients.length}</div>
            <div className="text-sm text-blue-600">{t('active_patients') || 'Active Patients'}</div>
          </div>
          
          <div className="bg-green-50 p-4 rounded-lg text-center">
            <Calendar className="h-8 w-8 text-green-600 mx-auto mb-2" />
            <div className="text-2xl font-bold text-green-600">{todaysSessions.length}</div>
            <div className="text-sm text-green-600">{t('todays_sessions') || "Today's Sessions"}</div>
          </div>
          
          <div className="bg-purple-50 p-4 rounded-lg text-center">
            <Activity className="h-8 w-8 text-purple-600 mx-auto mb-2" />
            <div className="text-2xl font-bold text-purple-600">4</div>
            <div className="text-sm text-purple-600">{t('departments') || 'Departments'}</div>
          </div>
          
          <div className="bg-orange-50 p-4 rounded-lg text-center">
            <CareIcon icon="l-user-md" className="h-8 w-8 text-orange-600 mx-auto mb-2" />
            <div className="text-2xl font-bold text-orange-600">12</div>
            <div className="text-sm text-orange-600">{t('medical_staff') || 'Medical Staff'}</div>
          </div>
        </div>
      </div>

      {/* Current Patients */}
      <div className="p-6 border-b border-gray-200">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">
          {t('current_patients') || 'Current Patients'}
        </h2>
        
        <div className="space-y-3">
          {climutuaPatients.map((patient, index) => (
            <div
              key={index}
              className="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
            >
              <div className="flex items-center space-x-3">
                <div className="w-10 h-10 bg-blue-100 rounded-full flex items-center justify-center">
                  <CareIcon icon="l-user" className="text-blue-600" />
                </div>
                <div>
                  <div className="font-medium text-gray-900">{patient.name}</div>
                  <div className="text-sm text-gray-500">
                    {t('age')}: {patient.age} • {patient.gender === 1 ? t('male') : t('female')}
                  </div>
                </div>
              </div>
              
              <div className="text-right">
                <div className="text-sm font-medium text-gray-900">
                  {patient.current_condition}
                </div>
                <div className="text-xs text-gray-500">
                  {patient.medical_history.join(', ')}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Today's Schedule */}
      <div className="p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">
          {t('todays_schedule') || "Today's Schedule"}
        </h2>
        
        <div className="space-y-3">
          {todaysSessions.map((session, index) => (
            <div
              key={index}
              className="flex items-center justify-between p-3 border border-gray-200 rounded-lg"
            >
              <div className="flex items-center space-x-3">
                <div className="text-sm font-medium text-gray-900">
                  {session.start_time} - {session.end_time}
                </div>
                <div className="text-sm text-gray-600">
                  {session.patient_name}
                </div>
              </div>
              
              <div className="text-right">
                <div className="text-sm font-medium text-gray-900">
                  {session.session_type}
                </div>
                <div className="text-xs text-gray-500">
                  {session.therapist_name} • {session.section}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default ClimutuaShowcase;