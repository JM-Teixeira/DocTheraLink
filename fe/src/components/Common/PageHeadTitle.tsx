import { useEffect } from "react";

export interface IPageTitleProps {
  title: string;
}

export default function PageTitle({ title }: IPageTitleProps) {
  useEffect(() => {
    const prevTitle = document.title;
    document.title = title + " | Klyniva";
    return () => {
      document.title = prevTitle;
    };
  }, [title]);

  return <></>;
}
