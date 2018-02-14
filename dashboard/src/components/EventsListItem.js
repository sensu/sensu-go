import React from "react";
import PropTypes from "prop-types";
import { createFragmentContainer, graphql } from "react-relay";
import moment from "moment";
import { withStyles } from "material-ui/styles";
import Checkbox from "material-ui/Checkbox";

const styles = theme => ({
  row: {
    width: "100%",
    paddingBottom: 16,
    borderBottomColor: theme.palette.divider,
    borderBottom: "1px solid",
    // TODO revist with typography
    fontFamily: "SF Pro Text",
  },
  checkbox: { display: "inline-block", verticalAlign: "top" },
  content: { display: "inline-block", padding: "16px 0 0" },
  command: { fontSize: "0.8125rem" },
});

class EventListItem extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
  };

  render() {
    const {
      classes,
      event: { entity, check, timestamp },
      ...other
    } = this.props;
    const time = moment(timestamp).fromNow();

    return (
      <div className={classes.row}>
        <div className={classes.checkbox}>
          <Checkbox />
        </div>
        <div className={classes.content}>
          <span>{entity.name}</span>
          <span>{check.config.name}</span>
          <div className={classes.command}>{check.config.command}</div>
          <div {...other} />
          <div>{time}</div>
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
    fragment EventRow_event on Event {
      ... on Event {
        timestamp
        check {
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
