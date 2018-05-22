import React from "react";
import PropTypes from "prop-types";
import { Query } from "react-apollo";
import gql from "graphql-tag";
import { withStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";

import Content from "/components/Content";
import AppContent from "/components/AppContent";
import NotFoundView from "/components/views/NotFoundView";

import EntitiesList from "/components/partials/EntitiesList";

// TODO: Abstract Headline component into the shared page layout component
const Headline = withStyles(theme => ({
  root: {
    display: "flex",
    alignContent: "center",
    paddingLeft: theme.spacing.unit,
    paddingRight: theme.spacing.unit,
    [theme.breakpoints.up("sm")]: {
      paddingLeft: 0,
      paddingRight: 0,
    },
    marginBottom: 16,
  },
}))(({ classes, ...props }) => <Content {...props} className={classes.root} />);

// TODO: Abstract Title component into the shared page layout component
const Title = withStyles(theme => ({
  root: {
    alignSelf: "flex-end",
    display: "none",
    flexGrow: 1,
    [theme.breakpoints.up("sm")]: {
      display: "flex",
    },
  },
}))(props => <Typography {...props} variant="headline" />);

class EntitiesContent extends React.PureComponent {
  static propTypes = {
    // eslint-disable-next-line react/no-unused-prop-types
    match: PropTypes.object.isRequired,
  };

  static query = gql`
    query EnvironmentViewEntitiesContentQuery(
      $environment: String!
      $organization: String!
      $sort: EntityListOrder = ID
    ) {
      environment(organization: $organization, environment: $environment) {
        ...EntitiesList_environment
      }
    }

    ${EntitiesList.fragments.environment}
  `;

  render() {
    return (
      <Query query={EntitiesContent.query} variables={this.props.match.params}>
        {({ data: { environment } = {}, loading, error, refetch }) => {
          // TODO: Connect this error handler to display a blocking error alert
          if (error) throw error;

          if (!environment && !loading) return <NotFoundView />;

          return (
            <AppContent>
              <Headline>
                <Title>Entities</Title>
              </Headline>
              <EntitiesList
                loading={loading}
                environment={environment}
                refetch={refetch}
              />
            </AppContent>
          );
        }}
      </Query>
    );
  }
}

export default EntitiesContent;
