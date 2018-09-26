/* eslint-disable react/sort-comp */

import React from "react";
import PropTypes from "prop-types";
import LinearProgress from "@material-ui/core/LinearProgress";

import Toast from "./Toast";
import ToastConnector from "./ToastConnector";

class PublishCheckStatusToast extends React.PureComponent {
  static propTypes = {
    mutation: PropTypes.object.isRequired,
    onClose: PropTypes.func.isRequired,
    checkName: PropTypes.string.isRequired,
    entityName: PropTypes.string,
    publish: PropTypes.bool,
  };

  static defaultProps = {
    entityName: undefined,
    publish: true,
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
    const { mutation, onClose, checkName, entityName } = this.props;
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

    const published = this.props.publish ? "Published" : "Unpublished";
    const publishing = this.props.publish ? "Publishing" : "Unpublishing";

    return (
      <ToastConnector>
        {({ addToast }) => (
          <Toast
            maxAge={loading ? undefined : 5000}
            variant={loading ? "info" : "success"}
            progress={loading && <LinearProgress />}
            message={
              loading ? (
                <span>
                  {publishing} {subject}.
                </span>
              ) : (
                <span>
                  {published} {subject}.{" "}
                </span>
              )
            }
            onClose={() => {
              onClose();

              if (loading) {
                const onMutationEnd = () =>
                  addToast(({ remove }) => (
                    <PublishCheckStatusToast {...this.props} onClose={remove} />
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

export default PublishCheckStatusToast;
