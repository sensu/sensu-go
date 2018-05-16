import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { Query } from "react-apollo";
import gql from "graphql-tag";
import { withStyles } from "material-ui/styles";
import Typography from "material-ui/Typography";
import Button from "material-ui/Button";

import AppContent from "/components/AppContent";
import EventsContainer from "/components/EventsContainer";
import SearchBox from "/components/SearchBox";
import Content from "/components/Content";
import NotFoundView from "/components/views/NotFoundView";

// If none given default expression is used.
const defaultExpression = "HasCheck && IsIncident";

class EventsContent extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    history: PropTypes.object.isRequired,
    match: PropTypes.object.isRequired,
    location: PropTypes.object.isRequired,
  };

  static styles = theme => ({
    headline: {
      display: "flex",
      alignContent: "center",
      paddingLeft: theme.spacing.unit,
      paddingRight: theme.spacing.unit,
      [theme.breakpoints.up("sm")]: {
        paddingLeft: 0,
        paddingRight: 0,
      },
    },
    searchBox: {
      width: "100%",
      [theme.breakpoints.up("sm")]: {
        marginLeft: theme.spacing.unit,
        width: "auto",
      },
    },
    title: {
      alignSelf: "flex-end",
      flexGrow: 1,
    },
    container: {
      marginTop: 10,
    },
    hiddenSmall: {
      display: "none",
      [theme.breakpoints.up("sm")]: {
        display: "flex",
      },
    },
  });

  static query = gql`
    query EnvironmentViewEventsContentQuery(
      $filter: String = "${defaultExpression}"
      $order: EventsListOrder = SEVERITY
      $environment: String!
      $organization: String!
    ) {
      environment(organization: $organization, environment: $environment) {
        ...EventsContainer_environment
      }
    }

    ${EventsContainer.fragments.environment}
  `;

  constructor(props) {
    super(props);

    const query = new URLSearchParams(props.location.search);

    let filterValue = query.filter;
    if (filterValue === undefined) {
      filterValue = defaultExpression;
    }
    this.state = { filterValue };
  }

  changeQuery = (key, val) => {
    const { location, history } = this.props;
    const query = new URLSearchParams(location.search);

    if (key === "filter") {
      this.setState({ filterValue: val });
    }
    query.set(key, val);

    history.push(`${location.pathname}?${query.toString()}`);
  };

  requerySearchBox = filterValue => {
    if (filterValue.length >= 10) {
      this.changeQuery("filter", filterValue);
    }
    this.setState({ filterValue });
  };

  render() {
    const { classes, match, location } = this.props;
    const query = new URLSearchParams(location.search);
    return (
      <Query
        query={EventsContent.query}
        fetchPolicy="cache-and-network"
        variables={{ ...match.params, filter: query.get("filter") }}
      >
        {({ data: { environment } = {}, loading, error, refetch }) => {
          if (error) throw error;

          if (!environment && !loading) return <NotFoundView />;

          return (
            <AppContent>
              <Content className={classes.headline}>
                <Typography
                  className={classnames(classes.title, classes.hiddenSmall)}
                  variant="headline"
                >
                  Events
                </Typography>
                <Button
                  className={classes.hiddenSmall}
                  onClick={() => refetch()}
                >
                  reload
                </Button>
                <SearchBox
                  className={classes.searchBox}
                  onChange={this.requerySearchBox}
                  value={this.state.filterValue}
                />
              </Content>
              <EventsContainer
                className={classes.container}
                onQueryChange={this.changeQuery}
                environment={environment}
                loading={loading}
              />
            </AppContent>
          );
        }}
      </Query>
    );
  }
}

export default withStyles(EventsContent.styles)(EventsContent);
