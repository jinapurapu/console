// This file is part of MinIO Console Server
// Copyright (c) 2022 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

import {
  ICertificateInfo,
  ISecurityContext,
  ITenantEncryptionResponse,
} from "../types";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import {
  containerForHeader,
  createTenantCommon,
  formFieldStyles,
  modalBasic,
  spacingUtils,
  tenantDetailsStyles,
  wizardCommon,
} from "../../Common/FormComponents/common/styleLibrary";
import React, { Fragment, useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { AppState } from "../../../../store";
import api from "../../../../common/api";
import { ErrorResponseHandler } from "../../../../common/types";

import FormSwitchWrapper from "../../Common/FormComponents/FormSwitchWrapper/FormSwitchWrapper";
import Grid from "@mui/material/Grid";
import FileSelector from "../../Common/FormComponents/FileSelector/FileSelector";
import InputBoxWrapper from "../../Common/FormComponents/InputBoxWrapper/InputBoxWrapper";
import RadioGroupSelector from "../../Common/FormComponents/RadioGroupSelector/RadioGroupSelector";
import { Button, DialogContentText } from "@mui/material";
import VisibilityOffIcon from "@mui/icons-material/VisibilityOff";
import RemoveRedEyeIcon from "@mui/icons-material/RemoveRedEye";
import { KeyPair } from "../ListTenants/utils";
import { clearValidationError } from "../utils";
import {
  commonFormValidation,
  IValidation,
} from "../../../../utils/validationFunctions";
import ConfirmDialog from "../../Common/ModalWrapper/ConfirmDialog";
import TLSCertificate from "../../Common/TLSCertificate/TLSCertificate";
import SectionTitle from "../../Common/SectionTitle";
import { setErrorSnackMessage } from "../../../../systemSlice";

interface ITenantEncryption {
  classes: any;
}

const styles = (theme: Theme) =>
  createStyles({
    ...tenantDetailsStyles,
    ...spacingUtils,
    bold: { fontWeight: "bold" },
    italic: { fontStyle: "italic" },
    fileItem: {
      marginRight: 10,
      display: "flex",
      "& div label": {
        minWidth: 50,
      },

      "@media (max-width: 900px)": {
        flexFlow: "column",
      },
    },
    formFieldRow: {
        marginBottom: ".8rem",
        "& .MuiInputLabel-root": {
          fontWeight: "normal",
        },
      },
      rightSpacer: {
        marginRight: 15,
      },
    ...containerForHeader(theme.spacing(4)),
    ...createTenantCommon,
    ...formFieldStyles,
    ...modalBasic,
    ...wizardCommon,
  });

const EditSecurityContext = () => {
  const dispatch = useDispatch();

  const tenant = useSelector(
    (state: AppState) => state.tenants.tenantDetails.tenantInfo
  );

  const [encryptionEnabled, setEncryptionEnabled] = useState<boolean>(false);
  const [encryptionType, setEncryptionType] = useState<string>("vault");
  const [replicas, setReplicas] = useState<string>("2");
  const [image, setImage] = useState<string>("");
  const [refreshEncryptionInfo, setRefreshEncryptionInfo] =
    useState<boolean>(false);
  const [securityContext, setSecurityContext] = useState<ISecurityContext>({
    fsGroup: "1000",
    runAsGroup: "1000",
    runAsNonRoot: true,
    runAsUser: "1000",
  });
  const [updatingEncryption, setUpdatingEncryption] = useState<boolean>(false);
  
  const [validationErrors, setValidationErrors] = useState<any>({});
  const cleanValidation = (fieldName: string) => {
    setValidationErrors(clearValidationError(validationErrors, fieldName));
  };
  const [confirmOpen, setConfirmOpen] = useState<boolean>(false);

  return (
    <React.Fragment>
            <Grid item xs={12}>
              <SectionTitle>Security Context for KES</SectionTitle>
            </Grid>
            <Grid item xs={12}>
              <div style={{ marginBottom: 10, display: "flex", }}>
                <div style={{   marginRight: 15,  marginBottom: ".8rem"}}  >
                  <InputBoxWrapper
                    type="number"
                    id="kes_securityContext_runAsUser"
                    name="kes_securityContext_runAsUser"
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      setSecurityContext({
                        ...securityContext,
                        runAsUser: e.target.value,
                      });
                    }}
                    label="Run As User"
                    value={securityContext.runAsUser}
                    required
                    error={
                      validationErrors["kes_securityContext_runAsUser"] || ""
                    }
                    min="0"
                  />
                </div>
                <div style={{ marginRight: 15, marginBottom: 10}} >
                  <InputBoxWrapper
                    type="number"
                    id="kes_securityContext_runAsGroup"
                    name="kes_securityContext_runAsGroup"
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      setSecurityContext({
                        ...securityContext,
                        runAsGroup: e.target.value,
                      });
                    }}
                    label="Run As Group"
                    value={securityContext.runAsGroup}
                    required
                    error={
                      validationErrors["kes_securityContext_runAsGroup"] || ""
                    }
                    min="0"
                  />
                </div>
                <div
                 style={{marginRight: 15, marginBottom: 10}}
                >
                  <InputBoxWrapper
                    type="number"
                    id="kes_securityContext_fsGroup"
                    name="kes_securityContext_fsGroup"
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      setSecurityContext({
                        ...securityContext,
                        fsGroup: e.target.value,
                      });
                    }}
                    label="FsGroup"
                    value={securityContext.fsGroup}
                    required
                    error={
                      validationErrors["kes_securityContext_fsGroup"] || ""
                    }
                    min="0"
                  />
                </div>
              </div>
            </Grid>
            <Grid item xs={12}>
              <FormSwitchWrapper
                value="kesSecurityContextRunAsNonRoot"
                id="kes_securityContext_runAsNonRoot"
                name="kes_securityContext_runAsNonRoot"
                checked={securityContext.runAsNonRoot}
                onChange={(e) => {
                  const targetD = e.target;
                  const checked = targetD.checked;
                  setSecurityContext({
                    ...securityContext,
                    runAsNonRoot: checked,
                  });
                }}
                label={"Do not run as Root"}
              />
            </Grid>
        </React.Fragment>
)
            }
export default withStyles(styles)(EditSecurityContext);
