/**
 * KlynivaTimetable Component
 * 
 * Enhanced timetable view for Klyniva healthcare management system
 * Integrates with the existing therapy system and provides comprehensive scheduling
 */

import React, { useState, useEffect } from 'react';
import { Calendar, Clock, User, MapPin, Plus, Filter } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import CareIcon from "@/CAREUI/icons/CareIcon";
import { sampleTimetableEntries } from "@/Utils/sampleData";

interface TimetableEntry {
  id?: string;
  patient_name: string;
  therapist_name: string;
  session_type: string;
  start_time: string;
  end_time: string;
  section: string;
  status: 'scheduled' | 'completed' | 'cancelled' | 'in_progress';
}

interface KlynivaTimetableProps {
  facilityId?: string;
  userRole?: string;
}

const KlynivaTimetable: React.FC<KlynivaTimetableProps> = ({ 
  facilityId, 
  userRole = 'staff' 
}) => {
  const { t } = useTranslation();
  const [selectedDate, setSelectedDate] = useState(new Date());
  const [entries, setEntries] = useState<TimetableEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedSection, setSelectedSection] = useState<string>('all');

  // Time slots from 8 AM to 6 PM
  const timeSlots = Array.from({ length: 20 }, (_, i) => {
    const hour = Math.floor(i / 2) + 8;
    const minute = (i % 2) * 30;
    return `${hour.toString().padStart(2, '0')}:${minute.toString().padStart(2, '0')}`;
  });

  const sections = [
    { id: 'all', name: t('all_sections') || 'All Sections' },
    { id: 'rehabilitation', name: t('rehabilitation') || 'Rehabilitation' },
    { id: 'cardiology', name: t('cardiology') || 'Cardiology' },
    { id: 'pulmonology', name: t('pulmonology') || 'Pulmonology' },
    { id: 'orthopedics', name: t('orthopedics') || 'Orthopedics' },
    { id: 'neurology', name: t('neurology') || 'Neurology' }
  ];

  useEffect(() => {
    loadTimetableData();
  }, [selectedDate, facilityId]);

  const loadTimetableData = async () => {
    setLoading(true);
    try {
      // For demo purposes, use sample data
      // In production, this would fetch from the API
      const sampleData = sampleTimetableEntries.map((entry, index) => ({
        ...entry,
        id: `entry-${index}`,
        status: entry.status as TimetableEntry['status']
      }));
      
      setEntries(sampleData);
    } catch (error) {
      console.error('Failed to load timetable data:', error);
    } finally {
      setLoading(false);
    }
  };

  const getFilteredEntries = () => {
    return entries.filter(entry => {
      if (selectedSection === 'all') return true;
      return entry.section.toLowerCase() === selectedSection;
    });
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'scheduled': return 'bg-blue-100 text-blue-800 border-blue-200';
      case 'completed': return 'bg-green-100 text-green-800 border-green-200';
      case 'cancelled': return 'bg-red-100 text-red-800 border-red-200';
      case 'in_progress': return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      default: return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  const formatDate = (date: Date) => {
    return date.toLocaleDateString('pt-PT', {
      weekday: 'long',
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  };

  const navigateDate = (direction: 'prev' | 'next') => {
    const newDate = new Date(selectedDate);
    newDate.setDate(newDate.getDate() + (direction === 'next' ? 1 : -1));
    setSelectedDate(newDate);
  };

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200">
      {/* Header */}
      <div className="p-6 border-b border-gray-200">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-2xl font-bold text-gray-900">
            {t('timetable') || 'Timetable'}
          </h2>
          
          <div className="flex items-center space-x-4">
            {/* Date Navigation */}
            <div className="flex items-center space-x-2 bg-gray-50 rounded-lg p-2">
              <button
                onClick={() => navigateDate('prev')}
                className="p-1 hover:bg-gray-200 rounded"
              >
                <CareIcon icon="l-angle-left" className="text-gray-600" />
              </button>
              
              <div className="flex items-center space-x-2 px-3">
                <Calendar className="h-5 w-5 text-gray-500" />
                <span className="font-medium text-gray-900 min-w-[200px] text-center">
                  {formatDate(selectedDate)}
                </span>
              </div>
              
              <button
                onClick={() => navigateDate('next')}
                className="p-1 hover:bg-gray-200 rounded"
              >
                <CareIcon icon="l-angle-right" className="text-gray-600" />
              </button>
            </div>

            {/* Section Filter */}
            <select
              value={selectedSection}
              onChange={(e) => setSelectedSection(e.target.value)}
              className="px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            >
              {sections.map(section => (
                <option key={section.id} value={section.id}>
                  {section.name}
                </option>
              ))}
            </select>
          </div>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-4 gap-4">
          <div className="bg-blue-50 p-3 rounded-lg">
            <div className="text-2xl font-bold text-blue-600">
              {entries.filter(e => e.status === 'scheduled').length}
            </div>
            <div className="text-sm text-blue-600">{t('scheduled') || 'Scheduled'}</div>
          </div>
          <div className="bg-yellow-50 p-3 rounded-lg">
            <div className="text-2xl font-bold text-yellow-600">
              {entries.filter(e => e.status === 'in_progress').length}
            </div>
            <div className="text-sm text-yellow-600">{t('in_progress') || 'In Progress'}</div>
          </div>
          <div className="bg-green-50 p-3 rounded-lg">
            <div className="text-2xl font-bold text-green-600">
              {entries.filter(e => e.status === 'completed').length}
            </div>
            <div className="text-sm text-green-600">{t('completed') || 'Completed'}</div>
          </div>
          <div className="bg-red-50 p-3 rounded-lg">
            <div className="text-2xl font-bold text-red-600">
              {entries.filter(e => e.status === 'cancelled').length}
            </div>
            <div className="text-sm text-red-600">{t('cancelled') || 'Cancelled'}</div>
          </div>
        </div>
      </div>

      {/* Timetable Content */}
      <div className="p-6">
        {loading ? (
          <div className="flex items-center justify-center py-12">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          </div>
        ) : (
          <div className="space-y-4">
            {getFilteredEntries().length === 0 ? (
              <div className="text-center py-12">
                <Calendar className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                <p className="text-lg font-medium text-gray-500 mb-2">
                  {t('no_sessions_scheduled') || 'No sessions scheduled'}
                </p>
                <p className="text-sm text-gray-400">
                  {t('no_therapy_sessions_today') || 'No therapy sessions scheduled for this day.'}
                </p>
              </div>
            ) : (
              getFilteredEntries().map((entry, index) => (
                <div
                  key={entry.id || index}
                  className="border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow"
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-4">
                      <div className="flex items-center space-x-2">
                        <Clock className="h-4 w-4 text-gray-500" />
                        <span className="font-medium text-gray-900">
                          {entry.start_time} - {entry.end_time}
                        </span>
                      </div>
                      
                      <div className="flex items-center space-x-2">
                        <User className="h-4 w-4 text-gray-500" />
                        <span className="text-gray-700">{entry.patient_name}</span>
                      </div>
                      
                      <div className="flex items-center space-x-2">
                        <MapPin className="h-4 w-4 text-gray-500" />
                        <span className="text-gray-700">{entry.section}</span>
                      </div>
                    </div>
                    
                    <div className="flex items-center space-x-3">
                      <span className={`px-2 py-1 rounded-full text-xs font-medium border ${getStatusColor(entry.status)}`}>
                        {t(entry.status) || entry.status}
                      </span>
                    </div>
                  </div>
                  
                  <div className="mt-2 text-sm text-gray-600">
                    <span className="font-medium">{entry.session_type}</span>
                    <span className="mx-2">•</span>
                    <span>{t('therapist') || 'Therapist'}: {entry.therapist_name}</span>
                  </div>
                </div>
              ))
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default KlynivaTimetable;