import { PatientModel } from "@/components/Patient/models";
import {
  Appointment,
  AppointmentRequest,
  AppointmentStatus,
  Board,
  Division,
  DivisionRequest,
  Shift,
  ShiftRequest,
} from "@/components/TimeBoard/types";

import { Type } from "@/Utils/request/api";
import { PaginatedResponse } from "@/Utils/request/types";

/** Time-board API (be/internal/modules/schedule/routes.go). */
const BoardRoutes = {
  board: {
    path: "/api/v1/schedule/board/",
    method: "GET",
    TRes: Type<Board>(),
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
    TBody: Type<{ status: AppointmentStatus }>(),
    TRes: Type<Appointment>(),
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
