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
  const linkURL = helpItemMarkdown.slice(
    helpItemMarkdown.indexOf("("),
    helpItemMarkdown.indexOf(")")
  );
  const imgURL = helpItemMarkdown.slice(0, helpItemMarkdown.indexOf("["));
  const altText = helpItemMarkdown.slice(
    helpItemMarkdown.indexOf("[") + 1,
    helpItemMarkdown.indexOf("]")
  );
  const markdownText = helpItemMarkdown.slice(helpItemMarkdown.indexOf("["));

  return (
    <Fragment>
      <Grid container alignItems={"center"} display={"flex"}>
        <Grid item xs={4} width={200}>
          <a href={linkURL}>
            <img src={imgURL} alt={altText} width="160" height="90" />
          </a>
        </Grid>
        <Grid
          item
          xs={8}
          width={400}
          style={{
            overflowWrap: "break-word",
          }}
        >
          <ReactMarkdown linkTarget="_blank" children={markdownText} />
        </Grid>
      </Grid>
    </Fragment>
  );
};

export default withStyles(styles)(HelpItem);
