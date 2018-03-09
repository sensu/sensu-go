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

import ResolveEventMutation from "../mutations/ResolveEventMutation";
import EventStatus from "./EventStatus";
import { TableListItem } from "./TableList";

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
    marginLeft: -12,
  },
  status: {
    display: "inline-block",
    verticalAlign: "top",
    padding: "14px 0",
  },
  disclosure: {
    color: theme.palette.action.active,
    marginLeft: 12,
    paddingTop: 14,
  },
  content: {
    width: "calc(100% - 96px)",
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

function fromNow(date) {
  const delta = new Date(date) - new Date();
  if (delta < 0) {
    return moment.duration(delta).humanize(true);
  }
  return "just now";
}

class EventListItem extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    checked: PropTypes.bool.isRequired,
    onChange: PropTypes.func.isRequired,

    relay: PropTypes.object.isRequired,
    event: PropTypes.shape({
      entity: PropTypes.shape({
        name: PropTypes.string.isRequired,
      }).isRequired,
      check: PropTypes.shape({
        name: PropTypes.string.isRequired,
        output: PropTypes.string.isRequired,
      }).isRequired,
      timestamp: PropTypes.string.isRequired,
    }).isRequired,
  };

  constructor(props) {
    super(props);
    this.fromNow = fromNow(props.event.timestamp);
  }

  state = { anchorEl: null };

  componentWillReceiveProps(nextProps) {
    if (this.props.event.timestamp !== nextProps.event.timestamp) {
      this.fromNow = fromNow(nextProps.event.timestamp);
    }
  }

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

  resolve = () => {
    const { relay, event } = this.props;
    ResolveEventMutation.commit(relay.environment, event.id, {});
    this.setState({ anchorEl: null });
  };

  render() {
    const { checked, classes, event: { entity, check }, onChange } = this.props;
    const { anchorEl } = this.state;
    const time = this.fromNow;

    return (
      <TableListItem selected={checked}>
        <div className={classes.checkbox}>
          <Checkbox onChange={onChange} checked={checked} />
        </div>
        <div className={classes.status}>
          <EventStatus status={check.status} />
        </div>
        <div className={classes.content}>
          <span className={classes.caption}>{entity.name}</span>
          <Chevron className={classes.chevron} />
          <span className={classes.caption}>{check.name}</span>
          <div className={classes.timeHolder}>
            Last occurred <em>&nbsp;{time}&nbsp;</em> and exited with status
            <em>&nbsp;{check.status}.</em>
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
            <MenuItem key={"resolve"} onClick={this.resolve}>
              Resolve
            </MenuItem>
          </Menu>
        </div>
      </TableListItem>
    );
  }
}

export default createFragmentContainer(
  withStyles(styles)(EventListItem),
  graphql`
    fragment EventsListItem_event on Event {
      ... on Event {
        id
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
