import careConfig from "@careConfig";
import {
  Popover,
  PopoverButton,
  PopoverPanel,
  Transition,
} from "@headlessui/react";
import { navigate } from "raviger";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

import Card from "@/CAREUI/display/Card";
import CareIcon from "@/CAREUI/icons/CareIcon";

import ButtonV2, { Cancel, Submit } from "@/components/Common/ButtonV2";
import GLocationPicker from "@/components/Common/GLocationPicker";
import Loading from "@/components/Common/Loading";
import Page from "@/components/Common/Page";
import SpokeFacilityEditor from "@/components/Facility/SpokeFacilityEditor";
import { DistrictModel, FacilityRequest } from "@/components/Facility/models";
import { PhoneNumberValidator } from "@/components/Form/FieldValidators";
import PhoneNumberFormField from "@/components/Form/FormFields/PhoneNumberFormField";
import RadioFormField from "@/components/Form/FormFields/RadioFormField";
import {
  MultiSelectFormField,
  SelectFormField,
} from "@/components/Form/FormFields/SelectFormField";
import TextAreaFormField from "@/components/Form/FormFields/TextAreaFormField";
import TextFormField from "@/components/Form/FormFields/TextFormField";
import { FieldChangeEvent } from "@/components/Form/FormFields/Utils";
import { FormAction } from "@/components/Form/Utils";

import useAppHistory from "@/hooks/useAppHistory";
import useAuthUser from "@/hooks/useAuthUser";

import { FACILITY_FEATURE_TYPES, FACILITY_TYPES } from "@/common/constants";
import {
  phonePreg,
  validateLatitude,
  validateLongitude,
  validatePincode,
} from "@/common/validation";

import { DraftSection, useAutoSaveReducer } from "@/Utils/AutoSave";
import * as Notification from "@/Utils/Notifications";
import routes from "@/Utils/request/api";
import request from "@/Utils/request/request";
import { RequestResult } from "@/Utils/request/types";
import useTanStackQueryInstead from "@/Utils/request/useQuery";
import {
  compareBy,
  getPincodeDetails,
  includesIgnoreCase,
  parsePhoneNumber,
} from "@/Utils/utils";

interface FacilityProps {
  facilityId?: string;
}

type FacilityForm = {
  facility_type?: number;
  name: string;
  state: number;
  district: number;
  features: number[];
  kasp_empanelled: string;
  address: string;
  latitude: string;
  longitude: string;
  pincode: string;
};

const initForm: FacilityForm = {
  facility_type: undefined,
  name: "",
  state: 0,
  district: 0,
  kasp_empanelled: "false",
  features: [],
  address: "",
  latitude: "",
  longitude: "",
  pincode: "",
};

const initError: Record<keyof FacilityForm, string> = Object.assign(
  {},
  ...Object.keys(initForm).map((k) => ({ [k]: "" })),
);

const initialState = {
  form: { ...initForm },
  errors: { ...initError },
};

const facilityCreateReducer = (state = initialState, action: FormAction) => {
  switch (action.type) {
    case "set_form":
      return { ...state, form: action.form };
    case "set_errors":
      return { ...state, errors: action.errors };
    case "set_state": {
      if (action.state) return action.state;
      return state;
    }
  }
};

/**
 * Safely converts a string coordinate value to a number or undefined.
 * @param value - The string value to convert
 * @returns number if valid, undefined if empty or invalid
 */
const convertCoordinate = (value: string): number | undefined => {
  if (!value || value.trim() === '') {
    return undefined;
  }
  const parsed = parseFloat(value);
  return isNaN(parsed) ? undefined : parsed;
};

export const FacilityCreate = (props: FacilityProps) => {
  const { t } = useTranslation();
  const { facilityId } = props;

  const [state, dispatch] = useAutoSaveReducer<FacilityForm>(
    facilityCreateReducer,
    initialState,
  );
  const [isLoading, setIsLoading] = useState(false);

  const [showAutoFilledPincode, setShowAutoFilledPincode] = useState(false);
  const [stateId, setStateId] = useState<number>();
  const [districtId, setDistrictId] = useState<number>();
  const { goBack } = useAppHistory();
  const headerText = !facilityId ? "Create Facility" : "Update Facility";
  const buttonText = !facilityId ? "Save Facility" : "Update Facility";

  const authUser = useAuthUser();
  useEffect(() => {
    if (
      authUser &&
      authUser.user_type !== "StateAdmin" &&
      authUser.user_type !== "DistrictAdmin" &&
      authUser.user_type !== "DistrictLabAdmin"
    ) {
      navigate("/facility");
      Notification.Error({
        msg: "You don't have permission to perform this action. Contact the admin",
      });
    }
  }, [authUser]);

  const {
    data: districtData,
    refetch: districtFetch,
    loading: isDistrictLoading,
  } = useTanStackQueryInstead(routes.getDistrictByState, {
    pathParams: {
      id: String(stateId),
    },
    prefetch: !!stateId,
  });



  const facilityQuery = useTanStackQueryInstead(routes.getPermittedFacility, {
    pathParams: {
      id: facilityId!,
    },
    prefetch: !!facilityId,
    onResponse: ({ res, data }) => {
      if (facilityId) {
        setIsLoading(true);
        if (res?.ok && data) {
          const formData = {
            facility_type: (typeof data.facility_type === 'number'
              ? data.facility_type
              : FACILITY_TYPES.find(ft => ft.text === String(data.facility_type))?.id) as number | undefined,
            name: data.name ? data.name : "",
            state: data.state ? data.state : 0,
            district: data.district ? data.district : 0,
            features: data.features || [],
            kasp_empanelled: "",
            address: data.address ? data.address : "",
            pincode: data.pincode ? data.pincode : "",
            latitude: data.latitude ? String(data.latitude) : "",
            longitude: data.longitude ? String(data.longitude) : "",
          };
          dispatch({ type: "set_form", form: formData });
          setStateId(data.state);
          setDistrictId(data.district);
        } else {
          navigate(`/facility/${facilityId}`);
        }
        setIsLoading(false);
      }
    },
  });

  const { data: stateData, loading: isStateLoading } = useTanStackQueryInstead(
    routes.statesList,
  );

  const handleChange = (e: FieldChangeEvent<unknown>) => {
    dispatch({
      type: "set_form",
      form: { ...state.form, [e.name]: e.value },
    });
  };

  const handleLocationChange = (location: google.maps.LatLng | undefined) => {
    if (location) {
      dispatch({
        type: "set_form",
        form: {
          ...state.form,
          latitude: location.lat().toFixed(7),
          longitude: location.lng().toFixed(7),
        },
      });
    }
  };

  const handlePincodeChange = async (e: FieldChangeEvent<string>) => {
    handleChange(e);
    // Note: Auto-fill from postal code is disabled for Portugal
    // Portugal uses XXXX-XXX format but doesn't have a public API for postal code lookup
    // Users must manually select state and district
  };

  const handleSelectCurrentLocation = (
    setCenter: (lat: number, lng: number) => void,
  ) => {
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition((position) => {
        dispatch({
          type: "set_form",
          form: {
            ...state.form,
            latitude: String(position.coords.latitude),
            longitude: String(position.coords.longitude),
          },
        });

        setCenter?.(position.coords.latitude, position.coords.longitude);
      });
    }
  };

  const validateForm = () => {
    const errors = { ...initError };
    let invalidForm = false;
    Object.keys(state.form).forEach((field) => {
      switch (field) {
        case "facility_type":
        case "name":
        case "address":
          if (!state.form[field]) {
            errors[field] = t("required");
            invalidForm = true;
          }
          return;

        case "district":
        case "state":
          if (!Number(state.form[field])) {
            errors[field] = t("required");
            invalidForm = true;
          }
          return;

        case "pincode":
          if (!validatePincode(state.form[field])) {
            errors[field] = t("invalid_pincode");
            invalidForm = true;
          }
          return;
        case "latitude":
          if (!!state.form.latitude && !validateLatitude(state.form[field])) {
            errors[field] = t("latitude_invalid");
            invalidForm = true;
          }
          return;
        case "longitude":
          if (!!state.form.longitude && !validateLongitude(state.form[field])) {
            errors[field] = t("longitude_invalid");
            invalidForm = true;
          }
          return;

        default:
          return;
      }
    });
    if (invalidForm) {
      dispatch({ type: "set_errors", errors });
      return false;
    }
    dispatch({ type: "set_errors", errors });
    return true;
  };

  const handleSubmit = async (e: any) => {
    e.preventDefault();
    const validated = validateForm();
    if (validated) {
      setIsLoading(true);
      
      // Ensure facility_type is a number, not a string (runtime safety check)
      const rawFacilityType = state.form.facility_type as any;
      const facilityType: number | undefined = typeof rawFacilityType === 'string'
        ? FACILITY_TYPES.find(ft => ft.text === rawFacilityType)?.id
        : rawFacilityType;
      
      const data: FacilityRequest = {
        facility_type: facilityType,
        name: state.form.name,
        district: state.form.district,
        state: state.form.state,
        address: state.form.address,
        features: state.form.features,
        pincode: state.form.pincode,
        latitude: convertCoordinate(state.form.latitude),
        longitude: convertCoordinate(state.form.longitude),
      };

      const { res, data: requestData } = facilityId
        ? await request(routes.updateFacility, {
          body: data,
          pathParams: {
            id: facilityId,
          },
        })
        : await request(routes.createFacility, {
          body: data,
        });

      if (res?.ok && requestData) {
        const id = requestData.id;
        dispatch({ type: "set_form", form: initForm });
        if (!facilityId) {
          Notification.Success({
            msg: "Facility added successfully",
          });
        } else {
          Notification.Success({
            msg: "Facility updated successfully",
          });
        }
        navigate(`/facility/${id}`);
      }
      setIsLoading(false);
    }
  };

  if (isLoading) {
    return <Loading />;
  }

  const field = (name: string) => {
    return {
      name,
      id: name,
      label: t(name),
      value: (state.form as any)[name],
      error: (state.errors as any)[name],
      onChange: handleChange,
    };
  };

  return (
    <Page
      title={headerText}
      crumbsReplacements={{
        [facilityId || "????"]: { name: state.form.name },
      }}
    >
      <Card className="mt-4">
        <div className="md:p-4">
          <form onSubmit={(e) => handleSubmit(e)}>
            <DraftSection
              handleDraftSelect={(newState: any) => {
                dispatch({ type: "set_state", state: newState });
                setStateId(newState.form.state);
                setDistrictId(newState.form.district);
              }}
              formData={state.form}
            />
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <SelectFormField
                {...field("facility_type")}
                required
                options={FACILITY_TYPES}
                optionLabel={(o) => o.text}
                optionValue={(o) => o.id}
              />
              <TextFormField
                {...field("name")}
                required
                label={t("facility_name")}
              />
              <MultiSelectFormField
                {...field("features")}
                placeholder={t("features")}
                options={FACILITY_FEATURE_TYPES}
                optionLabel={(o) => o.name}
                optionValue={(o) => o.id}
              />
              <div>
                <TextFormField
                  {...field("pincode")}
                  required
                  onChange={handlePincodeChange}
                />
                {showAutoFilledPincode && (
                  <div className="flex items-center gap-2 text-primary-500">
                    <CareIcon icon="l-check-circle" />
                    <span className="text-sm">
                      State and district auto-filled from pincode
                    </span>
                  </div>
                )}
              </div>
              <SelectFormField
                {...field("state")}
                required
                placeholder="Choose State"
                className={isStateLoading ? "animate-pulse" : ""}
                disabled={isStateLoading}
                options={stateData ? stateData.results : []}
                optionLabel={(o) => o.name}
                optionValue={(o) => o.id}
                onChange={(event) => {
                  handleChange(event);
                  if (!event) return;
                  setStateId(event.value);
                }}
              />
              <SelectFormField
                {...field("district")}
                placeholder="Choose District"
                required
                className={isDistrictLoading ? "animate-pulse" : ""}
                disabled={isDistrictLoading}
                options={districtData ? districtData : []}
                optionLabel={(o) => o.name}
                optionValue={(o) => o.id}
                onChange={(event) => {
                  handleChange(event);
                  if (!event) return;
                  setDistrictId(event.value);
                }}
              />
              <TextAreaFormField {...field("address")} required />
              {facilityId && (
                <div className="py-4 md:col-span-2">
                  <h4 className="mb-4">{t("spokes")}</h4>
                  <SpokeFacilityEditor
                    facility={{ ...facilityQuery.data, id: facilityId }}
                  />
                </div>
              )}
              {careConfig.kasp.enabled && (
                <RadioFormField
                  {...field("kasp_empanelled")}
                  label={`Is this facility ${careConfig.kasp.string} empanelled?`}
                  options={[true, false]}
                  optionLabel={(o) => (o ? "Yes" : "No")}
                  optionValue={(o) => String(o)}
                />
              )}
            </div>

            <div className="flex items-center gap-3">
              <TextFormField
                className="flex-1"
                {...field("latitude")}
                label={t("location")}
                placeholder="Latitude"
              />

              <div className="flex flex-col justify-center md:block">
                <Popover id="map-popover" className="relative">
                  <>
                    <PopoverButton>
                      <ButtonV2
                        circle
                        type="button"
                        id="facility-location-button"
                        className="tooltip p-2"
                      >
                        <CareIcon icon="l-map-marker" className="text-xl" />
                        <span className="tooltip-text tooltip-bottom">
                          Select location from map
                        </span>
                      </ButtonV2>
                    </PopoverButton>

                    <Transition
                      enter="transition ease-out duration-200"
                      enterFrom="opacity-0 translate-y-1"
                      enterTo="opacity-100 translate-y-0"
                      leave="transition ease-in duration-150"
                      leaveFrom="opacity-100 translate-y-0"
                      leaveTo="opacity-0 translate-y-1"
                    >
                      <PopoverPanel className="absolute -right-36 bottom-10 sm:-right-48">
                        <GLocationPicker
                          lat={Number(state.form.latitude)}
                          lng={Number(state.form.longitude)}
                          handleOnChange={handleLocationChange}
                          handleOnClose={() => null}
                          handleOnSelectCurrentLocation={
                            handleSelectCurrentLocation
                          }
                        />
                      </PopoverPanel>
                    </Transition>
                  </>
                </Popover>
              </div>
              <TextFormField
                className="flex-1"
                {...field("longitude")}
                label={<br />}
                placeholder="Longitude"
              />
            </div>
            <div className="mt-12 flex flex-col-reverse justify-end gap-3 sm:flex-row">
              <Cancel onClick={() => goBack()} />
              <Submit type="button" onClick={handleSubmit} label={buttonText} />
            </div>
          </form>
        </div>
      </Card>
    </Page>
  );
};
