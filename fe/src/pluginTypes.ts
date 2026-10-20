import { LazyExoticComponent } from "react";

import { INavItem } from "@/components/Common/Sidebar/Sidebar";
import { FacilityModel } from "@/components/Facility/models";
import { UserAssignedModel } from "@/components/Users/models";

import { AppRoutes } from "./Routers/AppRouter";
import { pluginMap } from "./pluginMap";

// Plugin system inherited from CARE (no plugins are enabled in Klyniva today).
export type AvailablePlugin = "@apps/care_livekit_fe";

export type AvailablePluginManifest = "@app-manifest/care_livekit_fe";

export type DoctorConnectButtonComponentType = React.FC<{
  user: UserAssignedModel;
}>;

export type ScribeComponentType = React.FC;
export type ManageFacilityOptionsComponentType = React.FC<{
  facility?: FacilityModel;
}>;

export type ExtendFacilityConfigureComponentType = React.FC<{
  facilityId: string;
}>;

// Define supported plugin components
export type SupportedPluginComponents = {
  DoctorConnectButtons: DoctorConnectButtonComponentType;
  Scribe: ScribeComponentType;
  ManageFacilityOptions: ManageFacilityOptionsComponentType;
  ExtendFacilityConfigure: ExtendFacilityConfigureComponentType;
};

// Create a type for lazy-loaded components
// eslint-disable-next-line @typescript-eslint/no-explicit-any
type LazyComponent<T extends React.FC<any>> = LazyExoticComponent<T>;

// Define PluginComponentMap with lazy-loaded components
export type PluginComponentMap = {
  [K in keyof SupportedPluginComponents]?: LazyComponent<
    SupportedPluginComponents[K]
  >;
};

type SupportedPluginExtensions = "DoctorConnectButtons";

export type PluginManifest = {
  plugin: string;
  routes: AppRoutes;
  extends: SupportedPluginExtensions[];
  components: PluginComponentMap;
  navItems: INavItem[];
};

// Create a type that ensures only available plugins can be used
export type EnabledPluginConfig = {
  plugin: string;
  manifestPath: AvailablePluginManifest;
  path: AvailablePlugin;
  manifest: Promise<PluginManifest>;
  // Components are a dictionary, with the key being the component name, and the value being the component type
  components: PluginComponentMap;
};

export { pluginMap };
