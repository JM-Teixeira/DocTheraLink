import { PatientModel } from "@/components/Patient/models";
import {
  Absence,
  AbsenceRequest,
  AbsenceResult,
  Appointment,
  AppointmentRequest,
  AppointmentStatus,
  Board,
  Division,
  DivisionRequest,
  MissCount,
  MissReport,
  MoveRequest,
  PatientHandover,
  PlanRescheduleRequest,
  PlanRescheduleResult,
  RangeView,
  SeriesRequest,
  SeriesResult,
  Shift,
  ShiftRequest,
  WaitingEntry,
  WaitingRequest,
  WaitingStatus,
  YearStats,
} from "@/components/TimeBoard/types";

import { Type } from "@/Utils/request/api";
import { PaginatedResponse } from "@/Utils/request/types";

/** Time-board API (be/internal/modules/schedule/routes.go). */
const BoardRoutes = {
  moveAppointment: {
    path: "/api/v1/schedule/appointments/{id}/move/",
    method: "POST",
    TBody: Type<MoveRequest>(),
    TRes: Type<Appointment>(),
  },
  seriesSessions: {
    path: "/api/v1/schedule/series/{id}/",
    method: "GET",
    TRes: Type<Appointment[]>(),
  },
  reschedulePlan: {
    path: "/api/v1/schedule/series/{id}/reschedule/",
    method: "POST",
    TBody: Type<PlanRescheduleRequest>(),
    TRes: Type<PlanRescheduleResult>(),
  },
  waiting: {
    path: "/api/v1/schedule/waiting/",
    method: "GET",
    TRes: Type<WaitingEntry[]>(),
  },
  addWaiting: {
    path: "/api/v1/schedule/waiting/",
    method: "POST",
    TBody: Type<WaitingRequest>(),
    TRes: Type<WaitingEntry>(),
  },
  updateWaiting: {
    path: "/api/v1/schedule/waiting/{id}/",
    method: "PATCH",
    TBody: Type<{
      status?: WaitingStatus;
      series_id?: string;
      urgent?: boolean;
      note?: string;
    }>(),
    TRes: Type<WaitingEntry>(),
  },
  missed: {
    path: "/api/v1/schedule/missed/",
    method: "GET",
    TRes: Type<MissReport>(),
  },
  patientMisses: {
    path: "/api/v1/schedule/missed/{patientId}/",
    method: "GET",
    TRes: Type<MissCount>(),
  },
  board: {
    path: "/api/v1/schedule/board/",
    method: "GET",
    TRes: Type<Board>(),
  },
  range: {
    path: "/api/v1/schedule/range/",
    method: "GET",
    TRes: Type<RangeView>(),
  },
  year: {
    path: "/api/v1/schedule/year/",
    method: "GET",
    TRes: Type<YearStats>(),
  },
  series: {
    path: "/api/v1/schedule/series/",
    method: "POST",
    TBody: Type<SeriesRequest>(),
    TRes: Type<SeriesResult>(),
  },
  cancelSeries: {
    path: "/api/v1/schedule/series/{id}/",
    method: "DELETE",
    TRes: Type<{ removed: number }>(),
  },
  repeatWeek: {
    path: "/api/v1/schedule/shifts/repeat-week/",
    method: "POST",
    TBody: Type<{ week_start: string; weeks: number }>(),
    TRes: Type<{ copied: number }>(),
  },
  absences: {
    path: "/api/v1/schedule/absences/",
    method: "GET",
    TRes: Type<Absence[]>(),
  },
  createAbsence: {
    path: "/api/v1/schedule/absences/",
    method: "POST",
    TBody: Type<AbsenceRequest>(),
    TRes: Type<AbsenceResult>(),
  },
  deleteAbsence: {
    path: "/api/v1/schedule/absences/{id}/",
    method: "DELETE",
    TRes: Type<void>(),
  },
  handover: {
    path: "/api/v1/schedule/handover/",
    method: "GET",
    TRes: Type<PatientHandover[]>(),
  },
  divisions: {
    path: "/api/v1/schedule/divisions/",
    method: "GET",
    TRes: Type<Division[]>(),
  },
  createDivision: {
    path: "/api/v1/schedule/divisions/",
    method: "POST",
    TBody: Type<DivisionRequest>(),
    TRes: Type<Division>(),
  },
  updateDivision: {
    path: "/api/v1/schedule/divisions/{id}/",
    method: "PUT",
    TBody: Type<DivisionRequest>(),
    TRes: Type<Division>(),
  },
  removeDivision: {
    path: "/api/v1/schedule/divisions/{id}/",
    method: "DELETE",
    TRes: Type<void>(),
  },
  createShift: {
    path: "/api/v1/schedule/shifts/",
    method: "POST",
    TBody: Type<ShiftRequest>(),
    TRes: Type<Shift>(),
  },
  updateShift: {
    path: "/api/v1/schedule/shifts/{id}/",
    method: "PUT",
    TBody: Type<ShiftRequest>(),
    TRes: Type<Shift>(),
  },
  deleteShift: {
    path: "/api/v1/schedule/shifts/{id}/",
    method: "DELETE",
    TRes: Type<void>(),
  },
  copyDay: {
    path: "/api/v1/schedule/shifts/copy/",
    method: "POST",
    TBody: Type<{ from_date: string; to_date: string }>(),
    TRes: Type<{ copied: number }>(),
  },
  book: {
    path: "/api/v1/schedule/appointments/",
    method: "POST",
    TBody: Type<AppointmentRequest>(),
    TRes: Type<Appointment>(),
  },
  setStatus: {
    path: "/api/v1/schedule/appointments/{id}/",
    method: "PATCH",
    TBody: Type<{
      status: AppointmentStatus;
      justified?: boolean;
      reason?: string;
    }>(),
    TRes: Type<Appointment & { warnings: string[] }>(),
  },
  unbook: {
    path: "/api/v1/schedule/appointments/{id}/",
    method: "DELETE",
    TRes: Type<void>(),
  },
  searchPatients: {
    path: "/api/v1/patient/",
    method: "GET",
    TRes: Type<PaginatedResponse<PatientModel>>(),
  },
} as const;

export default BoardRoutes;
