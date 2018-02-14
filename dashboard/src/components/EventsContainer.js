import React from "react";
import PropTypes from "prop-types";

import map from "lodash/map";
import get from "lodash/get";
import { createFragmentContainer, graphql } from "react-relay";
import { withStyles } from "material-ui/styles";

import checkboxIcon from "material-ui/Checkbox";
import arrowIcon from "material-ui-icons/ArrowDropDown";

import EventsListItem from "./EventsListItem";

const styles = theme => ({
  eventsContainer: {
    border: "1px solid",
    borderRadius: 2,
    borderColor: theme.palette.divider,
  },
  tableHeader: {
    padding: "20px 0 16px",
    backgroundColor: theme.palette.primary.light,
    color: theme.palette.primary.contrastText,
    // TODO revist with typography
    fontFamily: "SF Pro Text",
    textTransform: "uppercase",
    fontWeight: "bold",
    fontSize: "0.875rem",
    display: "flex",
    alignItems: "center",
  },
  tableHeaderButton: {
    marginLeft: 16,
    display: "flex",
  },
  arrow: { marginTop: -4 },
  checkbox: { marginTop: -4, width: 24, height: 24 },
});

class EventsContainer extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    viewer: PropTypes.shape({ checkEvents: PropTypes.object }).isRequired,
    DropdownArrow: PropTypes.func.isRequired,
    Checkbox: PropTypes.func.isRequired,
  };

  static defaultProps = {
    DropdownArrow: arrowIcon,
    Checkbox: checkboxIcon,
  };

  render() {
    const { classes, viewer, Checkbox, DropdownArrow } = this.props;
    const events = get(viewer, "events.edges", []);

    return (
      <div className={classes.eventsContainer}>
        <div className={classes.tableHeader}>
          <span className={classes.tableHeaderButton}>
            <Checkbox className={classes.checkbox} />
          </span>
          <span className={classes.tableHeaderButton}>
            Entity
            <DropdownArrow className={classes.arrow} />
          </span>
          <span className={classes.tableHeaderButton}>
            Check
            <DropdownArrow className={classes.arrow} />
          </span>
          <span className={classes.tableHeaderButton}>
            Status
            <DropdownArrow className={classes.arrow} />
          </span>
        </div>
        {map(events, (event, i) => (
          <EventsListItem key={i} event={event.node} />
        ))}
      </div>
    );
  }
}

export default createFragmentContainer(
  withStyles(styles)(EventsContainer),
  graphql`
    fragment EventList_viewer on Viewer {
      events(first: 1000) {
        edges {
          node {
            ...EventRow_event
          }
        }
        pageInfo {
          hasNextPage
        }
      }
    }
  `,
);
