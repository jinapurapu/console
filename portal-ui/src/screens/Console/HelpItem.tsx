// This file is part of MinIO Console Server
// Copyright (c) 2023 MinIO, Inc.
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

import React, { Fragment } from "react";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import Grid from "@mui/material/Grid";
import ReactMarkdown from "react-markdown";

import {
  containerForHeader,
  formFieldStyles,
  modalStyleUtils,
  spacingUtils,
} from "../Console/Common/FormComponents/common/styleLibrary";
import { wrap } from "lodash";
import { Typography } from "@mui/material";

const styles = (theme: Theme) =>
  createStyles({
    ...modalStyleUtils,
    ...formFieldStyles,
    ...spacingUtils,
    ...containerForHeader,
  });

interface IHelpItemProps {
  classes: any;
  helpItemMarkdown: string;
}

const HelpItem = ({ classes, helpItemMarkdown }: IHelpItemProps) => {
  return (
    <Fragment>
      <Grid container>
        <Grid
          item
          xs={12}
          width={500}
          style={{
            overflowWrap: "break-word",
          }}
        >
          <ReactMarkdown className={classes.helpMarkdown} linkTarget="_blank">
            {helpItemMarkdown}
          </ReactMarkdown>
        </Grid>
      </Grid>
    </Fragment>
  );
};

export default withStyles(styles)(HelpItem);
