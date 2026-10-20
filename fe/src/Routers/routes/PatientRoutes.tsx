import { Redirect } from "raviger";

import PatientDetail from "@/components/Patient/PatientDetail";
import PatientForm from "@/components/Patient/PatientForm";
import PatientList from "@/components/Patient/PatientList";

import { AppRoutes } from "@/Routers/AppRouter";

const PatientRoutes: AppRoutes = {
  "/patients": () => <PatientList />,
  "/patients/new": () => <PatientForm />,
  "/patients/:id": ({ id }) => <PatientDetail id={id} />,
  "/patients/:id/edit": ({ id }) => <PatientForm patientId={id} />,
  // Old CARE-style URLs
  "/patient/:id": ({ id }) => <Redirect to={`/patients/${id}`} />,
  "/facility/:facilityId/patient/:id": ({ id }) => (
    <Redirect to={`/patients/${id}`} />
  ),
};

export default PatientRoutes;
