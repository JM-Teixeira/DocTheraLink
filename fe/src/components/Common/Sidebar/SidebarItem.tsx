import { Link } from "raviger";
import React, { Ref, forwardRef } from "react";
import { useTranslation } from "react-i18next";

import CareIcon from "@/CAREUI/icons/CareIcon";

import useAppHistory from "@/hooks/useAppHistory";

export type SidebarIcon = React.ReactNode;

type SidebarItemProps = {
  id?: string;
  ref?: React.Ref<HTMLAnchorElement>;
  text: string;
  icon: SidebarIcon;
  onItemClick?: () => void;
  external?: true | undefined;
  badgeCount?: number | undefined;
  selected?: boolean | undefined;
  handleOverflow?: any;
} & ({ to?: string; do?: undefined } | { to?: string; do: () => void });

type SidebarItemBaseProps = SidebarItemProps & {
  shrinked?: boolean;
};

const SidebarItemBase = forwardRef<HTMLAnchorElement, SidebarItemBaseProps>(
  ({ shrinked, external, ...props }, ref) => {
    const { t } = useTranslation();
    const { resetHistory } = useAppHistory();

    return (
      <Link
        ref={ref}
        id={props?.id}
        className={`tooltip relative ml-1 mr-2 h-12 flex-1 cursor-pointer rounded-md py-1 font-medium text-gray-600 transition md:flex-none ${
          props.selected
            ? "bg-white text-green-800 shadow"
            : "font-normal" + (props.to || props.do ? " hover:bg-gray-200" : "")
        }`}
        target={external && "_blank"}
        rel={external && "noreferrer"}
        href={props.to ?? ""}
        onClick={() => {
          // On Review: Check if resetHistory is working as intended.
          props.do?.();
          props.onItemClick?.();
          resetHistory();
        }}
        onMouseEnter={() => {
          props.handleOverflow(true);
        }}
        onMouseLeave={() => {
          props.handleOverflow(false);
        }}
      >
        <span className={`tooltip-text tooltip-right ${!shrinked && "hidden"}`}>
          {t(props.text)}
        </span>
        {/* Same layout open and closed: the icon keeps its place (centred in
            the narrow bar) and the label fades out, so closing is as smooth
            as opening */}
        <div className="flex h-full items-center justify-start overflow-hidden pl-3 pr-2">
          <div className="flex-none text-lg">{props.icon}</div>
          <span
            aria-hidden={shrinked}
            className={`${
              shrinked
                ? "max-w-0 pl-0 opacity-0"
                : "max-w-[12rem] grow pl-4 opacity-100"
            } flex items-center overflow-hidden text-nowrap text-sm tracking-wide transition-all duration-300 ease-in-out`}
          >
            {t(props.text)}
          </span>
          {external && !shrinked && (
            <CareIcon icon="l-external-link-alt" className="text-lg" />
          )}
        </div>

        {!!props.badgeCount && (
          <span
            className={`absolute flex items-center justify-center bg-primary-500 font-semibold text-white ${
              shrinked
                ? "right-3 top-0.5 h-4 w-5 rounded-md text-xs"
                : "inset-y-0 right-4 my-auto h-6 rounded-md px-2 text-xs"
            } duration-400 z-10 animate-pulse transition-all ease-in-out`}
          >
            {props.badgeCount > 9 ? "9+" : props.badgeCount}
          </span>
        )}
      </Link>
    );
  },
);

export const SidebarItem = forwardRef(
  (props: SidebarItemProps, ref: Ref<HTMLAnchorElement>) => (
    <SidebarItemBase {...props} ref={ref} />
  ),
);

export const ShrinkedSidebarItem = forwardRef(
  (props: SidebarItemProps, ref: Ref<HTMLAnchorElement>) => (
    <SidebarItemBase shrinked ref={ref} {...props} />
  ),
);
