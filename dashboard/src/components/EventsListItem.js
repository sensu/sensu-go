import React from "react";
import PropTypes from "prop-types";

import { createFragmentContainer, graphql } from "react-relay";
import moment from "moment";
import { withStyles } from "material-ui/styles";
import Typography from "material-ui/Typography";

import Checkbox from "material-ui/Checkbox";
import chevronIcon from "material-ui-icons/ChevronRight";

import EventStatus from "./EventStatus";

const styles = theme => ({
  row: {
    display: "flex",
    width: "100%",
    borderBottomColor: theme.palette.divider,
    borderBottom: "1px solid",
    // TODO revist with typography
    fontFamily: "SF Pro Text",
  },
  checkbox: {
    display: "inline-block",
    verticalAlign: "top",
  },
  status: {
    display: "inline-block",
    verticalAlign: "top",
    padding: "14px 0",
  },
  content: {
    width: "100%",
    display: "inline-block",
    padding: 14,
  },
  command: { fontSize: "0.8125rem" },
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
    Chevron: PropTypes.func.isRequired,
  };

  static defaultProps = {
    Chevron: chevronIcon,
  };

  render() {
    const {
      classes,
      Chevron,
      event: { entity, check, timestamp },
      ...other
    } = this.props;
    const time = moment(timestamp).fromNow();

    return (
      <div className={classes.row}>
        <div className={classes.checkbox}>
          <Checkbox />
        </div>
        <div className={classes.status}>
          <EventStatus status={check.status} />
        </div>
        <div className={classes.content}>
          <span className={classes.caption}>{entity.name}</span>
          <Chevron className={classes.chevron} />
          <span className={classes.caption}>{check.config.name}</span>
          <div {...other} />
          <div className={classes.timeHolder}>
            Last ran<span className={classes.time}>&nbsp;{time}.</span>&nbsp;With
            an exit status of&nbsp;<span className={classes.time}>
              {check.status}.
            </span>
          </div>
          <Typography type="caption" className={classes.command}>
            {check.config.command}
          </Typography>
        </div>
      </div>
    );
  }
}

EventListItem.propTypes = {
  event: PropTypes.shape({
    entity: PropTypes.shape({ id: "" }).isRequired,
    check: PropTypes.shape({
      config: PropTypes.shape({ name: "", command: "" }),
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
          config {
            name
            command
          }
        }
        entity {
          name
        }
      }
    }
  `,
);
