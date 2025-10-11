/**
 * TimetableView Component
 * 
 * Displays a comprehensive timetable view for facility coordinators showing:
 * - Which therapists are treating which patients
 * - In which facility sections (pool, electric, massage, etc.)
 * - Time slots and session durations
 * - Color-coded sections for easy identification
 * 
 * Features:
 * - Date navigation
 * - Section filtering
 * - Therapist filtering
 * - Real-time updates
 * - Responsive design
 */

import React, { useState, useEffect } from 'react';
import { Calendar, Clock, User, MapPin, Filter, ChevronLeft, ChevronRight } from 'lucide-react';

interface TimetableEntry {
  id: string;
  patient_id: string;
  patient_name: string;
  therapist_id: number;
  therapist_name: string;
  section_id: string;
  section_name: string;
  section_color: string;
  start_time: string;
  end_time: string;
  status: string;
  duration: number;
}

interface TimetableResponse {
  date: string;
  entries: TimetableEntry[];
}

interface TimetableViewProps {
  facilityId: string;
  userRole: string;
}

const TimetableView: React.FC<TimetableViewProps> = ({ facilityId, userRole }) => {
  const [selectedDate, setSelectedDate] = useState(new Date());
  const [timetableData, setTimetableData] = useState<TimetableResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [selectedSection, setSelectedSection] = useState<string>('all');
  const [selectedTherapist, setSelectedTherapist] = useState<string>('all');
  const [sections, setSections] = useState<any[]>([]);
  const [therapists, setTherapists] = useState<any[]>([]);

  // Time slots for the timetable (8 AM to 6 PM)
  const timeSlots = Array.from({ length: 20 }, (_, i) => {
    const hour = Math.floor(i / 2) + 8;
    const minute = (i % 2) * 30;
    return `${hour.toString().padStart(2, '0')}:${minute.toString().padStart(2, '0')}`;
  });

  useEffect(() => {
    fetchTimetableData();
    fetchSections();
    fetchTherapists();
  }, [selectedDate, facilityId]);

  const fetchTimetableData = async () => {
    setLoading(true);
    try {
      const dateStr = selectedDate.toISOString().split('T')[0];
      const response = await fetch(
        `/api/v1/therapy/facilities/${facilityId}/timetable/?date=${dateStr}`,
        {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
          },
        }
      );
      
      if (response.ok) {
        const data = await response.json();
        setTimetableData(data);
      }
    } catch (error) {
      console.error('Failed to fetch timetable data:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchSections = async () => {
    try {
      const response = await fetch(`/api/v1/facility/${facilityId}/sections/`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
        },
      });
      
      if (response.ok) {
        const data = await response.json();
        setSections(data.results || []);
      }
    } catch (error) {
      console.error('Failed to fetch sections:', error);
    }
  };

  const fetchTherapists = async () => {
    try {
      const response = await fetch('/api/v1/users/?user_type=PhysioTherapist', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
        },
      });
      
      if (response.ok) {
        const data = await response.json();
        setTherapists(data.results || []);
      }
    } catch (error) {
      console.error('Failed to fetch therapists:', error);
    }
  };

  const navigateDate = (direction: 'prev' | 'next') => {
    const newDate = new Date(selectedDate);
    newDate.setDate(newDate.getDate() + (direction === 'next' ? 1 : -1));
    setSelectedDate(newDate);
  };

  const getFilteredEntries = () => {
    if (!timetableData) return [];
    
    return timetableData.entries.filter(entry => {
      const sectionMatch = selectedSection === 'all' || entry.section_id === selectedSection;
      const therapistMatch = selectedTherapist === 'all' || entry.therapist_id.toString() === selectedTherapist;
      return sectionMatch && therapistMatch;
    });
  };

  const getEntryPosition = (entry: TimetableEntry) => {
    const startTime = new Date(`2000-01-01T${entry.start_time}`);
    const startHour = startTime.getHours();
    const startMinute = startTime.getMinutes();
    
    // Calculate position from 8 AM
    const startPosition = ((startHour - 8) * 60 + startMinute) / 30;
    const duration = entry.duration / 30; // Convert minutes to 30-minute slots
    
    return {
      top: `${startPosition * 60}px`, // 60px per 30-minute slot
      height: `${duration * 60}px`,
    };
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return 'bg-green-100 border-green-300 text-green-800';
      case 'in_progress':
        return 'bg-blue-100 border-blue-300 text-blue-800';
      case 'cancelled':
        return 'bg-red-100 border-red-300 text-red-800';
      case 'no_show':
        return 'bg-orange-100 border-orange-300 text-orange-800';
      default:
        return 'bg-gray-100 border-gray-300 text-gray-800';
    }
  };

  const formatDate = (date: Date) => {
    return date.toLocaleDateString('en-US', {
      weekday: 'long',
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  const filteredEntries = getFilteredEntries();

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200">
      {/* Header */}
      <div className="p-6 border-b border-gray-200">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold text-gray-900">
            Facility Timetable
          </h2>
          
          {/* Date Navigation */}
          <div className="flex items-center space-x-4">
            <button
              onClick={() => navigateDate('prev')}
              className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
            >
              <ChevronLeft className="h-5 w-5" />
            </button>
            
            <div className="flex items-center space-x-2 bg-gray-50 px-4 py-2 rounded-lg">
              <Calendar className="h-5 w-5 text-gray-500" />
              <span className="font-medium text-gray-900">
                {formatDate(selectedDate)}
              </span>
            </div>
            
            <button
              onClick={() => navigateDate('next')}
              className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
            >
              <ChevronRight className="h-5 w-5" />
            </button>
          </div>
        </div>

        {/* Filters */}
        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-2">
            <Filter className="h-4 w-4 text-gray-500" />
            <span className="text-sm font-medium text-gray-700">Filters:</span>
          </div>
          
          <select
            value={selectedSection}
            onChange={(e) => setSelectedSection(e.target.value)}
            className="border border-gray-300 rounded-md px-3 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            <option value="all">All Sections</option>
            {sections.map((section) => (
              <option key={section.id} value={section.id}>
                {section.name}
              </option>
            ))}
          </select>
          
          <select
            value={selectedTherapist}
            onChange={(e) => setSelectedTherapist(e.target.value)}
            className="border border-gray-300 rounded-md px-3 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            <option value="all">All Therapists</option>
            {therapists.map((therapist) => (
              <option key={therapist.id} value={therapist.id}>
                {therapist.first_name} {therapist.last_name}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Timetable Grid */}
      <div className="p-6">
        <div className="relative">
          {/* Time Labels */}
          <div className="absolute left-0 top-0 w-16 h-full">
            {timeSlots.map((time, index) => (
              <div
                key={time}
                className="absolute text-xs text-gray-500 -translate-y-2"
                style={{ top: `${index * 60}px` }}
              >
                {time}
              </div>
            ))}
          </div>

          {/* Grid Lines */}
          <div className="ml-20 relative" style={{ height: `${timeSlots.length * 60}px` }}>
            {timeSlots.map((_, index) => (
              <div
                key={index}
                className="absolute w-full border-t border-gray-100"
                style={{ top: `${index * 60}px` }}
              />
            ))}

            {/* Timetable Entries */}
            {filteredEntries.map((entry) => (
              <div
                key={entry.id}
                className={`absolute left-0 right-0 mx-1 p-3 rounded-lg border-l-4 ${getStatusColor(entry.status)}`}
                style={{
                  ...getEntryPosition(entry),
                  borderLeftColor: entry.section_color,
                }}
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center space-x-2 mb-1">
                      <User className="h-4 w-4 flex-shrink-0" />
                      <span className="font-medium text-sm truncate">
                        {entry.patient_name}
                      </span>
                    </div>
                    
                    <div className="flex items-center space-x-2 mb-1">
                      <MapPin className="h-4 w-4 flex-shrink-0" />
                      <span className="text-xs text-gray-600 truncate">
                        {entry.section_name}
                      </span>
                    </div>
                    
                    <div className="flex items-center space-x-2">
                      <Clock className="h-4 w-4 flex-shrink-0" />
                      <span className="text-xs text-gray-600">
                        {entry.therapist_name}
                      </span>
                    </div>
                  </div>
                  
                  <div className="text-xs text-gray-500 ml-2">
                    {entry.duration}min
                  </div>
                </div>
              </div>
            ))}

            {/* Empty State */}
            {filteredEntries.length === 0 && (
              <div className="absolute inset-0 flex items-center justify-center">
                <div className="text-center text-gray-500">
                  <Calendar className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                  <p className="text-lg font-medium mb-2">No sessions scheduled</p>
                  <p className="text-sm">
                    No therapy sessions found for the selected date and filters.
                  </p>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Legend */}
      <div className="px-6 py-4 border-t border-gray-200 bg-gray-50">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-6">
            <span className="text-sm font-medium text-gray-700">Status:</span>
            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-2">
                <div className="w-3 h-3 bg-gray-200 rounded"></div>
                <span className="text-xs text-gray-600">Scheduled</span>
              </div>
              <div className="flex items-center space-x-2">
                <div className="w-3 h-3 bg-blue-200 rounded"></div>
                <span className="text-xs text-gray-600">In Progress</span>
              </div>
              <div className="flex items-center space-x-2">
                <div className="w-3 h-3 bg-green-200 rounded"></div>
                <span className="text-xs text-gray-600">Completed</span>
              </div>
              <div className="flex items-center space-x-2">
                <div className="w-3 h-3 bg-red-200 rounded"></div>
                <span className="text-xs text-gray-600">Cancelled</span>
              </div>
            </div>
          </div>
          
          <div className="text-sm text-gray-500">
            Total Sessions: {filteredEntries.length}
          </div>
        </div>
      </div>
    </div>
  );
};

export default TimetableView;