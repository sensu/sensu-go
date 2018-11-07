/* eslint-disable react/sort-comp */

import React from "react";
import PropTypes from "prop-types";
import LinearProgress from "@material-ui/core/LinearProgress";

import NamespaceLink from "/components/util/NamespaceLink";
import InlineLink from "/components/InlineLink";

import Toast from "./Toast";
import ToastConnector from "./ToastConnector";

class ExecuteCheckStatusToast extends React.PureComponent {
  static propTypes = {
    mutation: PropTypes.object.isRequired,
    onClose: PropTypes.func.isRequired,
    checkName: PropTypes.string.isRequired,
    entityName: PropTypes.string,
    namespace: PropTypes.string.isRequired,
  };

  static defaultProps = {
    entityName: undefined,
  };

  state = { loading: true };
  _willUnmount = false;

  componentDidMount() {
    this.props.mutation.then(
      () => !this._willUnmount && this.setState({ loading: false }),
      // TODO: Handle error cases
    );
  }

  componentWillUnmount() {
    // Prevent calling setState on unmounted component after mutation resolves
    this._willUnmount = true;
  }

  render() {
    const { mutation, onClose, checkName, entityName, namespace } = this.props;
    const { loading } = this.state;

    const subject = (
      <React.Fragment>
        <strong>{checkName}</strong>
        {entityName && (
          <span>
            {" "}
            on <strong>{entityName}</strong>
          </span>
        )}
      </React.Fragment>
    );

    return (
      <ToastConnector>
        {({ addToast }) => (
          <Toast
            maxAge={loading ? undefined : 5000}
            variant={loading ? "info" : "success"}
            progress={loading && <LinearProgress />}
            message={
              loading ? (
                <span>Executing {subject}.</span>
              ) : (
                <span>
                  Done executing {subject}.{" "}
                  <InlineLink
                    component={NamespaceLink}
                    namespace={namespace}
                    to={`/events?filter=${encodeURIComponent(
                      `Check.Name == '${checkName}'${
                        entityName ? ` && Entity.ID == '${entityName}'` : ""
                      }`,
                    )}`}
                  >
                    View&nbsp;events.
                  </InlineLink>
                </span>
              )
            }
            onClose={() => {
              onClose();

              if (loading) {
                const onMutationEnd = () =>
                  addToast(({ remove }) => (
                    <ExecuteCheckStatusToast {...this.props} onClose={remove} />
                  ));
                mutation.then(onMutationEnd, onMutationEnd);
              }
            }}
          />
        )}
      </ToastConnector>
    );
  }
}

export default ExecuteCheckStatusToast;
