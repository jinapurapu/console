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

import React, { Fragment, useEffect, useRef, useState } from "react";
import { Theme } from "@mui/material/styles";
import createStyles from "@mui/styles/createStyles";
import withStyles from "@mui/styles/withStyles";
import Grid from "@mui/material/Grid";

import {
  containerForHeader,
  formFieldStyles,
  modalStyleUtils,
  spacingUtils,
} from "../Console/Common/FormComponents/common/styleLibrary";
import { MenuItem, Paper, Tab, Tabs } from "@mui/material";
import HelpItem from "./HelpItem";
import { Box, Button, HelpIcon } from "mds";
import { useSelector } from "react-redux";
import { AppState } from "../../store";
import { TabPanel } from "../shared/tabs";

const styles = (theme: Theme) =>
  createStyles({
    ...modalStyleUtils,
    ...formFieldStyles,
    ...spacingUtils,
    ...containerForHeader,
  });

interface IHelpMenuProps {
  classes: any;
  helpName?: string;
}

const HelpMenu = ({ classes }: IHelpMenuProps) => {
  const [helpTab, setHelpTab] = useState<number>(0);
  const [helpItems, setHelpItems] = useState<string[]>([]);
  const [helpItemsVideo, setHelpItemsVideo] = useState<string[]>([]);
  const [helpItemsBlog, setHelpItemsBlog] = useState<string[]>([]);
  const [helpMenuOpen, setHelpMenuOpen] = useState<boolean>(false);

  const systemHelpName = useSelector(
    (state: AppState) => state.system.helpName
  );

  const helpTabName = useSelector(
    (state: AppState) => state.system.helpTabName
  );

  const toggleHelpMenu = () => {
    setHelpMenuOpen(!helpMenuOpen);
  };

  function useOutsideAlerter(ref: any) {
    useEffect(() => {
      function handleClickOutside(event: any) {
        if (ref.current && !ref.current.contains(event.target)) {
          setHelpMenuOpen(false);
        }
      }
      document.addEventListener("mousedown", handleClickOutside);
      return () => {
        document.removeEventListener("mousedown", handleClickOutside);
      };
    }, [ref]);
  }

  const wrapperRef = useRef(null);
  useOutsideAlerter(wrapperRef);

  useEffect(() => {
    assignHelpItems(systemHelpName);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [systemHelpName, helpTabName]);

  const assignHelpItems = (helpType: string) => {
    let helpTopics = require("../Console/helpTopics.json");
    if (helpType.includes("details")) {
      if (helpTopics[helpType + helpTabName]) {
        helpTopics[helpType + helpTabName]["docs"] &&
          setHelpItems(helpTopics[helpType + helpTabName]["docs"]);

        helpTopics[helpType + helpTabName]["blog"] &&
          setHelpItemsBlog(helpTopics[helpType + helpTabName]["blog"]);

        helpTopics[helpType + helpTabName]["video"] &&
          setHelpItemsVideo(helpTopics[helpType + helpTabName]["video"]);
      } else {
        setHelpItems(helpTopics["help"]["docs"]);
      }
    } else {
      if (helpTopics[helpType]) {
        helpTopics[helpType]["docs"] &&
          setHelpItems(helpTopics[helpType]["docs"]);

        helpTopics[helpType]["blog"] &&
          setHelpItemsBlog(helpTopics[helpType]["blog"]);

        helpTopics[helpType]["video"] &&
          setHelpItemsVideo(helpTopics[helpType]["video"]);
      } else {
        setHelpItems(helpTopics["help"]["docs"]);
      }
    }
  };

  const helpContent = (
    <React.Fragment>
      {helpItems &&
        helpItems.map((aHelpItem) => (
          <MenuItem value={aHelpItem} key={`help-item-${aHelpItem}`}>
            <HelpItem helpItemMarkdown={aHelpItem} />
          </MenuItem>
        ))}
    </React.Fragment>
  );
  const helpContentVideo = (
    <React.Fragment>
      {helpItemsVideo &&
        helpItemsVideo.map((aHelpItem) => (
          <MenuItem value={aHelpItem} key={`help-item-${aHelpItem}`}>
            <HelpItem helpItemMarkdown={aHelpItem} />
          </MenuItem>
        ))}
    </React.Fragment>
  );
  const helpContentBlog = (
    <React.Fragment>
      {helpItemsBlog &&
        helpItemsBlog.map((aHelpItem) => (
          <MenuItem value={aHelpItem} key={`help-item-${aHelpItem}`}>
            <HelpItem helpItemMarkdown={aHelpItem} />
          </MenuItem>
        ))}
    </React.Fragment>
  );
  function a11yProps(index: any) {
    return {
      id: `simple-tab-${index}`,
      "aria-controls": `simple-tabpanel-${index}`,
    };
  }

  return (
    <Fragment>
      {helpMenuOpen && (
        <div
          ref={wrapperRef}
          style={{
            position: "absolute",
            zIndex: "3",
            backgroundColor: "#d3d6d9",
            opacity: "0.95",
          }}
        >
          <Box>
            <Grid container alignItems={"center"}>
              <Grid item paddingLeft={3}>
                <HelpIcon />
              </Grid>
              <Grid item paddingLeft={3}>
                <Tabs
                  value={helpTab}
                  onChange={(e: React.ChangeEvent<{}>, newValue: number) => {
                    setHelpTab(newValue);
                  }}
                  indicatorColor="primary"
                  textColor="primary"
                  aria-label="cluster-tabs"
                  variant="scrollable"
                  scrollButtons="auto"
                >
                  {helpItems.length !== 0 && (
                    <Tab label="Documentation" {...a11yProps(0)} />
                  )}
                  {helpItemsVideo.length !== 0 && (
                    <Tab label="Video" {...a11yProps(1)} />
                  )}
                  {helpItemsBlog.length !== 0 && (
                    <Tab label="Blog" {...a11yProps(2)} />
                  )}
                </Tabs>
              </Grid>
            </Grid>
            <Paper>
              {helpItems.length !== 0 && (
                <TabPanel index={0} value={helpTab}>
                  {helpContent}
                </TabPanel>
              )}
              {helpItemsVideo.length !== 0 && (
                <TabPanel index={1} value={helpTab}>
                  {helpContentVideo}
                </TabPanel>
              )}
              {helpItemsBlog.length !== 0 && (
                <TabPanel
                  index={helpItemsVideo.length !== 0 ? 2 : 1}
                  value={helpTab}
                >
                  {helpContentBlog}
                </TabPanel>
              )}
            </Paper>
          </Box>
        </div>
      )}
      <Button
        id={systemHelpName ? systemHelpName : "help_button"}
        icon={<HelpIcon />}
        onClick={toggleHelpMenu}
      ></Button>
    </Fragment>
  );
};

export default withStyles(styles)(HelpMenu);
