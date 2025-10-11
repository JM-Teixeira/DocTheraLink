/**
 * PatientProfile Component
 * 
 * Comprehensive patient profile page for therapy patients showing:
 * - Patient demographics and contact information
 * - Treatment plan (P1) with goals and progress
 * - Session history and attendance tracking
 * - Notes and communications
 * - Stakeholder management
 * - Transfer capabilities
 * 
 * Features:
 * - Unique patient ID display (YYDDDRRRR format)
 * - Treatment plan management
 * - Session notes and progress tracking
 * - Stakeholder notifications
 * - Missed session tracking
 * - Transfer between therapists
 */

import React, { useState, useEffect } from 'react';
import { 
  User, 
  Calendar, 
  Phone, 
  MapPin, 
  FileText, 
  Users, 
  AlertCircle,
  Edit3,
  Plus,
  Clock,
  Target,
  Activity,
  MessageSquare,
  UserPlus
} from 'lucide-react';

interface TherapyPatient {
  id: string;
  unique_id: string;
  name: string;
  date_of_birth?: string;
  gender?: string;
  phone_number?: string;
  emergency_contact?: string;
  address?: string;
  medical_history?: string;
  admission_date: string;
  treatment_type: string;
  referring_doctor?: string;
  diagnosis_notes?: string;
  treatment_goals?: string;
  total_sessions: number;
  completed_sessions: number;
  missed_sessions: number;
  is_active: boolean;
  discharge_date?: string;
  discharge_notes?: string;
  facility: any;
  assigned_therapist?: any;
  assigned_doctor?: any;
  treatment_plan?: any;
  stakeholders: any[];
}

interface PatientProfileProps {
  patientId: string;
  currentUserRole: string;
  currentUserId: number;
}

const PatientProfile: React.FC<PatientProfileProps> = ({ 
  patientId, 
  currentUserRole, 
  currentUserId 
}) => {
  const [patient, setPatient] = useState<TherapyPatient | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('overview');
  const [notes, setNotes] = useState<any[]>([]);
  const [sessions, setSessions] = useState<any[]>([]);
  const [showAddNote, setShowAddNote] = useState(false);
  const [showAddStakeholder, setShowAddStakeholder] = useState(false);

  useEffect(() => {
    fetchPatientData();
    fetchPatientNotes();
    fetchPatientSessions();
  }, [patientId]);

  const fetchPatientData = async () => {
    try {
      const response = await fetch(`/api/v1/therapy/patients/${patientId}/`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
        },
      });
      
      if (response.ok) {
        const data = await response.json();
        setPatient(data);
      }
    } catch (error) {
      console.error('Failed to fetch patient data:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchPatientNotes = async () => {
    try {
      const response = await fetch(`/api/v1/therapy/patients/${patientId}/notes/`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
        },
      });
      
      if (response.ok) {
        const data = await response.json();
        setNotes(data.results || []);
      }
    } catch (error) {
      console.error('Failed to fetch patient notes:', error);
    }
  };

  const fetchPatientSessions = async () => {
    try {
      const response = await fetch(`/api/v1/therapy/sessions/?patient_id=${patientId}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
        },
      });
      
      if (response.ok) {
        const data = await response.json();
        setSessions(data.results || []);
      }
    } catch (error) {
      console.error('Failed to fetch patient sessions:', error);
    }
  };

  const updateMissedSessions = async (missedCount: number) => {
    try {
      const response = await fetch(
        `/api/v1/therapy/patients/${patientId}/missed-sessions/`,
        {
          method: 'PATCH',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
          },
          body: JSON.stringify({ missed_sessions: missedCount }),
        }
      );
      
      if (response.ok) {
        fetchPatientData(); // Refresh patient data
      }
    } catch (error) {
      console.error('Failed to update missed sessions:', error);
    }
  };

  const addPatientNote = async (noteData: any) => {
    try {
      const response = await fetch(`/api/v1/therapy/patients/${patientId}/notes/`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
        },
        body: JSON.stringify(noteData),
      });
      
      if (response.ok) {
        fetchPatientNotes(); // Refresh notes
        setShowAddNote(false);
      }
    } catch (error) {
      console.error('Failed to add patient note:', error);
    }
  };

  const calculateAge = (dateOfBirth: string) => {
    const today = new Date();
    const birthDate = new Date(dateOfBirth);
    let age = today.getFullYear() - birthDate.getFullYear();
    const monthDiff = today.getMonth() - birthDate.getMonth();
    
    if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birthDate.getDate())) {
      age--;
    }
    
    return age;
  };

  const getAttendanceRate = () => {
    if (!patient || patient.total_sessions === 0) return 0;
    return Math.round((patient.completed_sessions / patient.total_sessions) * 100);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  const canEditPatient = () => {
    return ['Doctor', 'PhysioTherapist', 'LocalBodyAdmin', 'DistrictAdmin', 'StateAdmin'].includes(currentUserRole);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  if (!patient) {
    return (
      <div className="text-center py-12">
        <AlertCircle className="h-12 w-12 mx-auto mb-4 text-red-500" />
        <p className="text-lg font-medium text-gray-900 mb-2">Patient not found</p>
        <p className="text-gray-500">The requested patient could not be found.</p>
      </div>
    );
  }

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      {/* Patient Header */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <div className="flex items-start justify-between">
          <div className="flex items-start space-x-4">
            <div className="w-16 h-16 bg-primary-100 rounded-full flex items-center justify-center">
              <User className="h-8 w-8 text-primary-600" />
            </div>
            
            <div>
              <div className="flex items-center space-x-3 mb-2">
                <h1 className="text-2xl font-bold text-gray-900">{patient.name}</h1>
                <span className="px-3 py-1 bg-primary-100 text-primary-800 text-sm font-medium rounded-full">
                  ID: {patient.unique_id}
                </span>
                {!patient.is_active && (
                  <span className="px-3 py-1 bg-red-100 text-red-800 text-sm font-medium rounded-full">
                    Discharged
                  </span>
                )}
              </div>
              
              <div className="flex items-center space-x-6 text-sm text-gray-600">
                {patient.date_of_birth && (
                  <div className="flex items-center space-x-1">
                    <Calendar className="h-4 w-4" />
                    <span>Age {calculateAge(patient.date_of_birth)}</span>
                  </div>
                )}
                
                {patient.phone_number && (
                  <div className="flex items-center space-x-1">
                    <Phone className="h-4 w-4" />
                    <span>{patient.phone_number}</span>
                  </div>
                )}
                
                <div className="flex items-center space-x-1">
                  <MapPin className="h-4 w-4" />
                  <span>{patient.facility?.name}</span>
                </div>
              </div>
            </div>
          </div>
          
          {canEditPatient() && (
            <button className="flex items-center space-x-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors">
              <Edit3 className="h-4 w-4" />
              <span>Edit Patient</span>
            </button>
          )}
        </div>

        {/* Quick Stats */}
        <div className="grid grid-cols-4 gap-4 mt-6">
          <div className="bg-blue-50 p-4 rounded-lg">
            <div className="flex items-center space-x-2 mb-2">
              <Activity className="h-5 w-5 text-blue-600" />
              <span className="text-sm font-medium text-blue-900">Total Sessions</span>
            </div>
            <div className="text-2xl font-bold text-blue-600">{patient.total_sessions}</div>
          </div>
          
          <div className="bg-green-50 p-4 rounded-lg">
            <div className="flex items-center space-x-2 mb-2">
              <Target className="h-5 w-5 text-green-600" />
              <span className="text-sm font-medium text-green-900">Completed</span>
            </div>
            <div className="text-2xl font-bold text-green-600">{patient.completed_sessions}</div>
          </div>
          
          <div className="bg-orange-50 p-4 rounded-lg">
            <div className="flex items-center space-x-2 mb-2">
              <AlertCircle className="h-5 w-5 text-orange-600" />
              <span className="text-sm font-medium text-orange-900">Missed</span>
            </div>
            <div className="text-2xl font-bold text-orange-600">{patient.missed_sessions}</div>
          </div>
          
          <div className="bg-purple-50 p-4 rounded-lg">
            <div className="flex items-center space-x-2 mb-2">
              <Clock className="h-5 w-5 text-purple-600" />
              <span className="text-sm font-medium text-purple-900">Attendance</span>
            </div>
            <div className="text-2xl font-bold text-purple-600">{getAttendanceRate()}%</div>
          </div>
        </div>
      </div>

      {/* Tab Navigation */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200">
        <div className="border-b border-gray-200">
          <nav className="flex space-x-8 px-6">
            {[
              { id: 'overview', label: 'Overview', icon: User },
              { id: 'treatment', label: 'Treatment Plan', icon: FileText },
              { id: 'sessions', label: 'Sessions', icon: Calendar },
              { id: 'notes', label: 'Notes', icon: MessageSquare },
              { id: 'stakeholders', label: 'Stakeholders', icon: Users },
            ].map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center space-x-2 py-4 px-1 border-b-2 font-medium text-sm ${
                  activeTab === tab.id
                    ? 'border-primary-500 text-primary-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                <tab.icon className="h-4 w-4" />
                <span>{tab.label}</span>
              </button>
            ))}
          </nav>
        </div>

        {/* Tab Content */}
        <div className="p-6">
          {activeTab === 'overview' && (
            <div className="grid grid-cols-2 gap-6">
              <div>
                <h3 className="text-lg font-medium text-gray-900 mb-4">Patient Information</h3>
                <div className="space-y-3">
                  <div>
                    <label className="text-sm font-medium text-gray-500">Admission Date</label>
                    <p className="text-gray-900">{formatDate(patient.admission_date)}</p>
                  </div>
                  
                  <div>
                    <label className="text-sm font-medium text-gray-500">Treatment Type</label>
                    <p className="text-gray-900">{patient.treatment_type}</p>
                  </div>
                  
                  {patient.referring_doctor && (
                    <div>
                      <label className="text-sm font-medium text-gray-500">Referring Doctor</label>
                      <p className="text-gray-900">{patient.referring_doctor}</p>
                    </div>
                  )}
                  
                  {patient.emergency_contact && (
                    <div>
                      <label className="text-sm font-medium text-gray-500">Emergency Contact</label>
                      <p className="text-gray-900">{patient.emergency_contact}</p>
                    </div>
                  )}
                </div>
              </div>
              
              <div>
                <h3 className="text-lg font-medium text-gray-900 mb-4">Medical Information</h3>
                <div className="space-y-3">
                  {patient.diagnosis_notes && (
                    <div>
                      <label className="text-sm font-medium text-gray-500">Diagnosis</label>
                      <p className="text-gray-900">{patient.diagnosis_notes}</p>
                    </div>
                  )}
                  
                  {patient.medical_history && (
                    <div>
                      <label className="text-sm font-medium text-gray-500">Medical History</label>
                      <p className="text-gray-900">{patient.medical_history}</p>
                    </div>
                  )}
                  
                  {patient.treatment_goals && (
                    <div>
                      <label className="text-sm font-medium text-gray-500">Treatment Goals</label>
                      <p className="text-gray-900">{patient.treatment_goals}</p>
                    </div>
                  )}
                </div>
              </div>
            </div>
          )}

          {activeTab === 'treatment' && (
            <div>
              {patient.treatment_plan ? (
                <div className="space-y-6">
                  <div className="flex items-center justify-between">
                    <h3 className="text-lg font-medium text-gray-900">Treatment Plan (P1)</h3>
                    {canEditPatient() && (
                      <button className="flex items-center space-x-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors">
                        <Edit3 className="h-4 w-4" />
                        <span>Edit Plan</span>
                      </button>
                    )}
                  </div>
                  
                  <div className="bg-gray-50 p-6 rounded-lg">
                    <div className="grid grid-cols-2 gap-6">
                      <div>
                        <h4 className="font-medium text-gray-900 mb-2">Diagnosis</h4>
                        <p className="text-gray-700">{patient.treatment_plan.diagnosis}</p>
                      </div>
                      
                      <div>
                        <h4 className="font-medium text-gray-900 mb-2">Treatment Goals</h4>
                        <p className="text-gray-700">{patient.treatment_plan.treatment_goals}</p>
                      </div>
                    </div>
                    
                    <div className="mt-4">
                      <h4 className="font-medium text-gray-900 mb-2">Treatment Plan</h4>
                      <p className="text-gray-700 whitespace-pre-wrap">{patient.treatment_plan.treatment_plan}</p>
                    </div>
                  </div>
                </div>
              ) : (
                <div className="text-center py-12">
                  <FileText className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                  <p className="text-lg font-medium text-gray-500 mb-2">No Treatment Plan</p>
                  <p className="text-gray-400 mb-4">This patient doesn't have a treatment plan yet.</p>
                  {canEditPatient() && (
                    <button className="flex items-center space-x-2 mx-auto px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors">
                      <Plus className="h-4 w-4" />
                      <span>Create Treatment Plan</span>
                    </button>
                  )}
                </div>
              )}
            </div>
          )}

          {activeTab === 'notes' && (
            <div>
              <div className="flex items-center justify-between mb-6">
                <h3 className="text-lg font-medium text-gray-900">Patient Notes</h3>
                <button
                  onClick={() => setShowAddNote(true)}
                  className="flex items-center space-x-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
                >
                  <Plus className="h-4 w-4" />
                  <span>Add Note</span>
                </button>
              </div>
              
              {notes.length === 0 ? (
                <div className="text-center py-12">
                  <MessageSquare className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                  <p className="text-lg font-medium text-gray-500 mb-2">No Notes</p>
                  <p className="text-gray-400">No notes have been added for this patient yet.</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {notes.map((note) => (
                    <div key={note.id} className="bg-gray-50 p-4 rounded-lg">
                      <div className="flex items-start justify-between mb-2">
                        <div>
                          <h4 className="font-medium text-gray-900">{note.title}</h4>
                          <p className="text-sm text-gray-500">
                            By {note.author.first_name} {note.author.last_name} • {formatDate(note.created_date)}
                          </p>
                        </div>
                        <span className={`px-2 py-1 text-xs font-medium rounded-full ${
                          note.priority === 'urgent' ? 'bg-red-100 text-red-800' :
                          note.priority === 'high' ? 'bg-orange-100 text-orange-800' :
                          'bg-gray-100 text-gray-800'
                        }`}>
                          {note.priority}
                        </span>
                      </div>
                      <p className="text-gray-700 whitespace-pre-wrap">{note.content}</p>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {activeTab === 'stakeholders' && (
            <div>
              <div className="flex items-center justify-between mb-6">
                <h3 className="text-lg font-medium text-gray-900">Patient Stakeholders</h3>
                <button
                  onClick={() => setShowAddStakeholder(true)}
                  className="flex items-center space-x-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
                >
                  <UserPlus className="h-4 w-4" />
                  <span>Add Stakeholder</span>
                </button>
              </div>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {patient.stakeholders.map((stakeholder) => (
                  <div key={stakeholder.id} className="bg-gray-50 p-4 rounded-lg">
                    <div className="flex items-center justify-between">
                      <div>
                        <h4 className="font-medium text-gray-900">
                          {stakeholder.user.first_name} {stakeholder.user.last_name}
                        </h4>
                        <p className="text-sm text-gray-500">{stakeholder.role}</p>
                      </div>
                      {stakeholder.is_primary && (
                        <span className="px-2 py-1 bg-primary-100 text-primary-800 text-xs font-medium rounded-full">
                          Primary
                        </span>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default PatientProfile;