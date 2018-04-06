import React from "react";
import PropTypes from "prop-types";

import { graphql } from "react-relay";
import { routerShape, matchShape } from "found";
import { withStyles } from "material-ui/styles";
import Typography from "material-ui/Typography";

import AppContent from "../components/AppContent";
import EventsContainer from "../components/EventsContainer";
import SearchBox from "../components/SearchBox";

// If none given default expression is used.
const defaultExpression = "HasCheck && IsIncident";

class EventsPage extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    router: routerShape.isRequired,
    match: matchShape.isRequired,
  };

  static styles = theme => ({
    headline: {
      display: "flex",
      justifyContent: "space-between",
      alignContent: "center",
    },
    searchBox: {
      width: "100%",
      marginLeft: theme.spacing.unit,
      marginRight: theme.spacing.unit,
      [theme.breakpoints.up("sm")]: {
        width: "auto",
        margin: 0,
      },
    },
    title: {
      alignSelf: "flex-end",
      display: "none",
      [theme.breakpoints.up("sm")]: {
        display: "flex",
      },
    },
    container: {
      marginTop: 10,
    },
  });

  static query = graphql`
    query EventsPageQuery(
      $filter: String = "HasCheck && IsIncident"
      $order: EventsListOrder = SEVERITY
      $environment: String!
      $organization: String!
    ) {
      environment(organization: $organization, environment: $environment) {
        ...EventsContainer_environment
      }
    }
  `;

  constructor(props) {
    super(props);

    let filterValue = props.match.location.query.filter;
    if (filterValue === undefined) {
      filterValue = defaultExpression;
    }
    this.state = { filterValue };
  }

  changeQuery = (key, val) => {
    const { match, router } = this.props;
    const query = new URLSearchParams(match.location.query);

    if (key === "filter") {
      this.setState({ filterValue: val });
    }
    query.set(key, val);

    router.push(`${match.location.pathname}?${query.toString()}`);
  };

  requerySearchBox = filterValue => {
    if (filterValue.length >= 10) {
      this.changeQuery("filter", filterValue);
    }
    this.setState({ filterValue });
  };

  render() {
    const { classes, ...props } = this.props;
    return (
      <AppContent>
        <div>
          <div className={classes.headline}>
            <Typography className={classes.title} variant="headline">
              Events
            </Typography>
            <SearchBox
              className={classes.searchBox}
              onChange={this.requerySearchBox}
              value={this.state.filterValue}
            />
          </div>
          <EventsContainer
            className={classes.container}
            onQueryChange={this.changeQuery}
            {...props}
          />
        </div>
      </AppContent>
    );
  }
}

export default withStyles(EventsPage.styles)(EventsPage);
