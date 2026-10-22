import { RefObject, useContext, useEffect } from "react";

import { cn } from "@/lib/utils";

import PageHelp from "@/components/Common/PageHelp";
import PageTitle, { PageTitleProps } from "@/components/Common/PageTitle";
import { SidebarShrinkContext } from "@/components/Common/Sidebar/Sidebar";

interface PageProps extends PageTitleProps {
  children: React.ReactNode | React.ReactNode[];
  options?: React.ReactNode | React.ReactNode[];
  changePageMetadata?: boolean;
  className?: string;
  noImplicitPadding?: boolean;
  ref?: RefObject<HTMLDivElement>;
  /**
   * If true, the sidebar will be collapsed when mounted, and restored to original state when unmounted.
   * @default false
   **/
  collapseSidebar?: boolean;
  hideTitleOnPage?: boolean;
  /** Topic of the "?" help pop-up next to the title (locale help__<topic>) */
  help?: string;
}

export default function Page(props: PageProps) {
  const sidebar = useContext(SidebarShrinkContext);

  useEffect(() => {
    if (!props.collapseSidebar) return;

    sidebar.setShrinked(true);
    return () => {
      sidebar.setShrinked(sidebar.shrinked);
    };
  }, [props.collapseSidebar]);

  let padding = "";
  if (!props.noImplicitPadding) {
    if (!props.hideBack || props.componentRight) padding = "py-3 md:px-6";
    else padding = "px-6 py-5";
  }

  return (
    <div className={cn(padding, props.className)} ref={props.ref}>
      <div className="mb-4 flex flex-col justify-between gap-2 px-3 md:flex-row md:items-center md:gap-6 md:px-0">
        <PageTitle
          changePageMetadata={props.changePageMetadata}
          title={props.title}
          breadcrumbs={props.breadcrumbs}
          backUrl={props.backUrl}
          hideBack={props.hideBack}
          componentRight={props.componentRight}
          crumbsReplacements={props.crumbsReplacements}
          focusOnLoad={props.focusOnLoad}
          onBackClick={props.onBackClick}
          isInsidePage={true}
          hideTitleOnPage={props.hideTitleOnPage}
        />
        {props.help && (
          <div className="md:mr-auto md:-ml-4">
            <PageHelp topic={props.help} title={props.title} />
          </div>
        )}
        {props.options}
      </div>
      {props.children}
    </div>
  );
}
