import React from "react";
import PropTypes from "prop-types";
import moment from "moment";

import { createFragmentContainer, graphql } from "react-relay";
import { withStyles } from "material-ui/styles";
import Typography from "material-ui/Typography";
import Menu, { MenuItem } from "material-ui/Menu";
import Button from "material-ui/ButtonBase";

import Checkbox from "material-ui/Checkbox";
import Chevron from "material-ui-icons/ChevronRight";
import Disclosure from "material-ui-icons/MoreVert";

import EventStatus from "./EventStatus";

const styles = theme => ({
  row: {
    display: "flex",
    width: "100%",
    borderColor: theme.palette.divider,
    border: "1px solid",
    borderTop: "none",
    color: theme.palette.text.primary,
  },
  checkbox: {
    display: "inline-block",
    verticalAlign: "top",
    marginLeft: 4,
  },
  status: {
    display: "inline-block",
    verticalAlign: "top",
    padding: "14px 0",
  },
  disclosure: {
    marginRight: 4,
    paddingTop: 14,
    color: theme.palette.action.active,
  },
  content: {
    width: "calc(100% - 104px)",
    display: "inline-block",
    padding: 14,
  },
  command: {
    fontSize: "0.8125rem",
    whiteSpace: "nowrap",
    overflow: "hidden",
    textOverflow: "ellipsis",
  },
  chevron: {
    verticalAlign: "top",
    marginTop: -2,
    color: theme.palette.primary.light,
  },
  timeHolder: {
    width: "100%",
    display: "flex",
    fontSize: "0.8125rem",
    margin: "4px 0 6px",
  },
  pipe: { marginTop: -4 },
});

class EventListItem extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
  };

  state = { anchorEl: null };

  onClose = () => {
    this.setState({ anchorEl: null });
  };

  handleClick = event => {
    this.setState({ anchorEl: event.currentTarget });
  };

  silenceEntity = entity => () => {
    // eslint-disable-next-line
    console.info("entity", entity);
    this.setState({ anchorEl: null });
  };

  silenceCheck = check => () => {
    // eslint-disable-next-line
    console.info("check", check);
    this.setState({ anchorEl: null });
  };

  resolve = event => () => {
    // eslint-disable-next-line
    console.info("event", event);
    this.setState({ anchorEl: null });
  };

  render() {
    const {
      classes,
      event: { entity, check, timestamp },
      ...other
    } = this.props;
    const { anchorEl } = this.state;
    const time = moment(timestamp).fromNow();

    return (
      <Typography component="div" className={classes.row}>
        <div className={classes.checkbox}>
          <Checkbox />
        </div>
        <div className={classes.status}>
          <EventStatus status={check.status} />
        </div>
        <div className={classes.content}>
          <span className={classes.caption}>{entity.name}</span>
          <Chevron className={classes.chevron} />
          <span className={classes.caption}>{check.name}</span>
          <div {...other} />
          <div className={classes.timeHolder}>
            Last ran<span className={classes.time}>&nbsp;{time}.</span>&nbsp;With
            an exit status of&nbsp;<span className={classes.time}>
              {check.status}.
            </span>
          </div>
          <Typography type="caption" className={classes.command}>
            {check.output}
          </Typography>
        </div>
        <div className={classes.disclosure}>
          <Button onClick={this.handleClick}>
            <Disclosure />
          </Button>
          {/* TODO give these functionality, pass correct value */}
          <Menu
            anchorEl={anchorEl}
            open={Boolean(anchorEl)}
            onClose={this.onClose}
            id="silenceButton"
          >
            <MenuItem
              key={"silence-Entity"}
              onClick={this.silenceEntity("entity")}
            >
              Silence Entity
            </MenuItem>
            <MenuItem
              key={"silence-Check"}
              onClick={this.silenceCheck("entity")}
            >
              Silence Check
            </MenuItem>
            <MenuItem key={"resolve"} onClick={this.resolve("event")}>
              Resolve
            </MenuItem>
          </Menu>
        </div>
      </Typography>
    );
  }
}

EventListItem.propTypes = {
  event: PropTypes.shape({
    entity: PropTypes.shape({ id: "" }).isRequired,
    check: PropTypes.shape({
      name: "",
      output: "",
    }).isRequired,
    timestamp: PropTypes.string.isRequired,
  }).isRequired,
};

export default createFragmentContainer(
  withStyles(styles)(EventListItem),
  graphql`
    fragment EventsListItem_event on Event {
      ... on Event {
        timestamp
        check {
          status
          name
          output
        }
        entity {
          name
        }
      }
    }
  `,
);
