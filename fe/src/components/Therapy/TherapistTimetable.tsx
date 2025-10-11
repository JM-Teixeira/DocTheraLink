/**
 * TherapistTimetable Component
 * 
 * Personal timetable view for individual therapists showing:
 * - Their scheduled patients for the day
 * - Session times and locations
 * - Patient information and treatment notes
 * - Quick access to session management
 * 
 * Features:
 * - Personal schedule view
 * - Session status updates
 * - Patient quick access
 * - Notes and treatment tracking
 */

import React, { useState, useEffect } from 'react';
import { Calendar, Clock, User, MapPin, FileText, Edit3, CheckCircle, XCircle } from 'lucide-react';

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

interface TherapistTimetableProps {
  therapistId: number;
  currentUserId: number;
}

const TherapistTimetable: React.FC<TherapistTimetableProps> = ({ 
  therapistId, 
  currentUserId 
}) => {
  const [selectedDate, setSelectedDate] = useState(new Date());
  const [sessions, setSessions] = useState<TimetableEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedSession, setSelectedSession] = useState<string | null>(null);

  useEffect(() => {
    fetchTherapistSchedule();
  }, [selectedDate, therapistId]);

  const fetchTherapistSchedule = async () => {
    setLoading(true);
    try {
      const dateStr = selectedDate.toISOString().split('T')[0];
      const response = await fetch(
        `/api/v1/therapy/therapists/${therapistId}/timetable/?date=${dateStr}`,
        {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
          },
        }
      );
      
      if (response.ok) {
        const data = await response.json();
        setSessions(data.entries || []);
      }
    } catch (error) {
      console.error('Failed to fetch therapist schedule:', error);
    } finally {
      setLoading(false);
    }
  };

  const updateSessionStatus = async (sessionId: string, status: string) => {
    try {
      const response = await fetch(
        `/api/v1/therapy/sessions/${sessionId}/notes/`,
        {
          method: 'PATCH',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
          },
          body: JSON.stringify({ status }),
        }
      );
      
      if (response.ok) {
        // Refresh the schedule
        fetchTherapistSchedule();
      }
    } catch (error) {
      console.error('Failed to update session status:', error);
    }
  };

  const formatTime = (timeString: string) => {
    return new Date(`2000-01-01T${timeString}`).toLocaleTimeString('en-US', {
      hour: 'numeric',
      minute: '2-digit',
      hour12: true,
    });
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className="h-5 w-5 text-green-600" />;
      case 'cancelled':
      case 'no_show':
        return <XCircle className="h-5 w-5 text-red-600" />;
      default:
        return <Clock className="h-5 w-5 text-gray-400" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return 'bg-green-50 border-green-200';
      case 'in_progress':
        return 'bg-blue-50 border-blue-200';
      case 'cancelled':
        return 'bg-red-50 border-red-200';
      case 'no_show':
        return 'bg-orange-50 border-orange-200';
      default:
        return 'bg-white border-gray-200';
    }
  };

  const isCurrentSession = (session: TimetableEntry) => {
    const now = new Date();
    const sessionStart = new Date(`${selectedDate.toDateString()} ${session.start_time}`);
    const sessionEnd = new Date(`${selectedDate.toDateString()} ${session.end_time}`);
    
    return now >= sessionStart && now <= sessionEnd;
  };

  const sortedSessions = sessions.sort((a, b) => 
    a.start_time.localeCompare(b.start_time)
  );

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200">
      {/* Header */}
      <div className="p-6 border-b border-gray-200">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold text-gray-900">
            My Schedule
          </h2>
          
          <div className="flex items-center space-x-2 bg-gray-50 px-4 py-2 rounded-lg">
            <Calendar className="h-5 w-5 text-gray-500" />
            <span className="font-medium text-gray-900">
              {selectedDate.toLocaleDateString('en-US', {
                weekday: 'long',
                month: 'long',
                day: 'numeric',
              })}
            </span>
          </div>
        </div>

        {/* Quick Stats */}
        <div className="grid grid-cols-4 gap-4">
          <div className="bg-blue-50 p-3 rounded-lg">
            <div className="text-2xl font-bold text-blue-600">
              {sortedSessions.length}
            </div>
            <div className="text-sm text-blue-600">Total Sessions</div>
          </div>
          
          <div className="bg-green-50 p-3 rounded-lg">
            <div className="text-2xl font-bold text-green-600">
              {sortedSessions.filter(s => s.status === 'completed').length}
            </div>
            <div className="text-sm text-green-600">Completed</div>
          </div>
          
          <div className="bg-orange-50 p-3 rounded-lg">
            <div className="text-2xl font-bold text-orange-600">
              {sortedSessions.filter(s => s.status === 'scheduled').length}
            </div>
            <div className="text-sm text-orange-600">Upcoming</div>
          </div>
          
          <div className="bg-purple-50 p-3 rounded-lg">
            <div className="text-2xl font-bold text-purple-600">
              {sortedSessions.reduce((total, session) => total + session.duration, 0)}
            </div>
            <div className="text-sm text-purple-600">Total Minutes</div>
          </div>
        </div>
      </div>

      {/* Sessions List */}
      <div className="p-6">
        {sortedSessions.length === 0 ? (
          <div className="text-center py-12">
            <Calendar className="h-12 w-12 mx-auto mb-4 text-gray-300" />
            <p className="text-lg font-medium text-gray-500 mb-2">
              No sessions scheduled
            </p>
            <p className="text-sm text-gray-400">
              You have no therapy sessions scheduled for today.
            </p>
          </div>
        ) : (
          <div className="space-y-4">
            {sortedSessions.map((session) => (
              <div
                key={session.id}
                className={`p-4 rounded-lg border-2 transition-all duration-200 ${
                  getStatusColor(session.status)
                } ${
                  isCurrentSession(session) ? 'ring-2 ring-primary-500' : ''
                } ${
                  selectedSession === session.id ? 'ring-2 ring-blue-500' : ''
                }`}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-4">
                    {/* Time */}
                    <div className="flex flex-col items-center min-w-0">
                      <div className="text-lg font-semibold text-gray-900">
                        {formatTime(session.start_time)}
                      </div>
                      <div className="text-sm text-gray-500">
                        {session.duration}min
                      </div>
                    </div>

                    {/* Patient Info */}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center space-x-2 mb-2">
                        <User className="h-5 w-5 text-gray-400" />
                        <span className="font-medium text-gray-900 truncate">
                          {session.patient_name}
                        </span>
                        {isCurrentSession(session) && (
                          <span className="px-2 py-1 bg-primary-100 text-primary-800 text-xs font-medium rounded-full">
                            Current
                          </span>
                        )}
                      </div>
                      
                      <div className="flex items-center space-x-2">
                        <MapPin className="h-4 w-4 text-gray-400" />
                        <span 
                          className="text-sm text-gray-600 px-2 py-1 rounded-full text-white"
                          style={{ backgroundColor: session.section_color }}
                        >
                          {session.section_name}
                        </span>
                      </div>
                    </div>
                  </div>

                  {/* Actions */}
                  <div className="flex items-center space-x-2">
                    {getStatusIcon(session.status)}
                    
                    {/* Quick Actions */}
                    <div className="flex items-center space-x-1">
                      <button
                        onClick={() => setSelectedSession(
                          selectedSession === session.id ? null : session.id
                        )}
                        className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
                        title="View Details"
                      >
                        <FileText className="h-4 w-4 text-gray-500" />
                      </button>
                      
                      <button
                        onClick={() => {/* Navigate to patient page */}}
                        className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
                        title="Edit Session"
                      >
                        <Edit3 className="h-4 w-4 text-gray-500" />
                      </button>
                    </div>

                    {/* Status Actions */}
                    {session.status === 'scheduled' && (
                      <div className="flex items-center space-x-1">
                        <button
                          onClick={() => updateSessionStatus(session.id, 'completed')}
                          className="px-3 py-1 bg-green-600 text-white text-sm rounded-md hover:bg-green-700 transition-colors"
                        >
                          Complete
                        </button>
                        <button
                          onClick={() => updateSessionStatus(session.id, 'no_show')}
                          className="px-3 py-1 bg-orange-600 text-white text-sm rounded-md hover:bg-orange-700 transition-colors"
                        >
                          No Show
                        </button>
                      </div>
                    )}
                  </div>
                </div>

                {/* Expanded Details */}
                {selectedSession === session.id && (
                  <div className="mt-4 pt-4 border-t border-gray-200">
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <h4 className="font-medium text-gray-900 mb-2">Session Details</h4>
                        <div className="space-y-1 text-sm text-gray-600">
                          <div>Patient ID: {session.patient_id}</div>
                          <div>Duration: {session.duration} minutes</div>
                          <div>Status: {session.status}</div>
                        </div>
                      </div>
                      
                      <div>
                        <h4 className="font-medium text-gray-900 mb-2">Quick Actions</h4>
                        <div className="space-y-2">
                          <button className="w-full text-left px-3 py-2 text-sm bg-gray-50 hover:bg-gray-100 rounded-md transition-colors">
                            View Patient Profile
                          </button>
                          <button className="w-full text-left px-3 py-2 text-sm bg-gray-50 hover:bg-gray-100 rounded-md transition-colors">
                            Add Session Notes
                          </button>
                          <button className="w-full text-left px-3 py-2 text-sm bg-gray-50 hover:bg-gray-100 rounded-md transition-colors">
                            Update Treatment Plan
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default TherapistTimetable;