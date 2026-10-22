import AboutPage from "@/components/About/AboutPage";
import AddPerson from "@/components/Staff/AddPerson";
import StaffList from "@/components/Staff/StaffList";
import UserHome from "@/components/Users/UserHome";

import { AppRoutes } from "@/Routers/AppRouter";

const UserRoutes: AppRoutes = {
  "/about": () => <AboutPage />,
  "/users": () => <StaffList />,
  "/users/add": () => <AddPerson />,
  "/users/:username": ({ username }) => (
    <UserHome username={username} tab={"profile"} />
  ),
  "/users/:username/:tab": ({ username, tab }) => (
    <UserHome username={username} tab={tab} />
  ),
  "/user/:tab": ({ tab }) => <UserHome tab={tab} />,
};

export default UserRoutes;
