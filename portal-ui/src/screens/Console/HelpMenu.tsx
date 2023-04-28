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
import { MenuItem, Select } from "@mui/material";
import HelpItem from "./HelpItem";
import { Box, Button, HelpIcon } from "mds";
import { useSelector } from "react-redux";
import { AppState } from "../../store";
import VerticalTabs from "./Common/VerticalTabs/VerticalTabs";

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

const HelpMenu = ({ classes, helpName }: IHelpMenuProps) => {
  const [helpItems, setHelpItems] = useState<string[]>([]);
  const [helpItemsVideo, setHelpItemsVideo] = useState<string[]>([]);
  const [helpItemsBlog, setHelpItemsBlog] = useState<string[]>([]);
  const [helpMenuOpen, setHelpMenuOpen] = useState<boolean>(false);

  const systemHelpName = useSelector(
    (state: AppState) => state.system.helpName
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
    let helpTopics = require("../Console/helpTopics.json");
    !helpName
      ? systemHelpName
        ? setHelpItems(helpTopics[systemHelpName])
        : setHelpItems(helpTopics[0])
      : setHelpItems(helpTopics[helpName]);
  }, [systemHelpName]);

  useEffect(() => {
    !helpName
      ? systemHelpName
        ? assignHelpItems(systemHelpName)
        : assignHelpItems("help")
      : assignHelpItems(helpName);
  }, [systemHelpName]);

  const assignHelpItems = (boop: string) => {
    let helpTopics = require("../Console/helpTopics.json");
    helpTopics[boop]["docs"] && setHelpItems(helpTopics[boop]["docs"]);
    helpTopics[boop]["blog"] && setHelpItemsBlog(helpTopics[boop]["blog"]);
    helpTopics[boop]["video"] && setHelpItemsVideo(helpTopics[boop]["video"]);
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
            Help Menu
            <VerticalTabs>
              {{
                tabConfig: {
                  label: "Documents",
                  disabled: helpItems.length === 0,
                },
                content: helpContent,
              }}
              {{
                tabConfig: {
                  label: "Video",
                  disabled: helpItemsVideo.length === 0,
                },
                content: helpContentVideo,
              }}
              {{
                tabConfig: {
                  label: "Blogs",
                  disabled: helpItemsBlog.length === 0,
                },
                content: helpContentBlog,
              }}
            </VerticalTabs>
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
