import React, { Fragment } from "react";
import { IAM_PAGES } from "../../../common/SecureComponent/permissions";
import { BackLink } from "mds";
import { useNavigate } from "react-router-dom";
import PageHeaderWrapper from "../Common/PageHeaderWrapper/PageHeaderWrapper";
import HelpMenu from "../HelpMenu";
import { useSelector } from "react-redux";
import { AppState } from "../../../store";

const GroupDetailsHeader = () => {
  const navigate = useNavigate();
  const systemHelpName = useSelector(
    (state: AppState) => state.system.helpName
  );
  return (
    <PageHeaderWrapper
      label={
        <Fragment>
          <BackLink
            label={"Groups"}
            onClick={() => navigate(IAM_PAGES.GROUPS)}
          />
        </Fragment>
      }
      actions={<HelpMenu helpName={`group_details`} />}
    />
  );
};

export default GroupDetailsHeader;
