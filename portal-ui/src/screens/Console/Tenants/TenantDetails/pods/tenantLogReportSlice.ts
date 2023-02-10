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
import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import { TenantLogReportMessage } from "../../../HealthInfo/types";

export interface TenantLogReportState {
  message: TenantLogReportMessage;
}

const initialState: TenantLogReportState = {
  message: {} as TenantLogReportMessage,
};

export const tenantLogReportSlice = createSlice({
  name: "trace",
  initialState,
  reducers: {
    tenantLogReportInfoMessageReceived: (
      state,
      action: PayloadAction<TenantLogReportMessage>
    ) => {
      state.message = action.payload;
    },
    tenantLogReportResetMessage: (state) => {
      state.message = {} as TenantLogReportMessage;
    },
  },
});

export const {
  tenantLogReportInfoMessageReceived,
  tenantLogReportResetMessage,
} = tenantLogReportSlice.actions;

export default tenantLogReportSlice.reducer;
