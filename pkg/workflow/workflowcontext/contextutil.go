/*
Package workflowcontext contains utility functions for populating workflow context specific data
in a context.Context.

Data appropriate for the context includes anything that cannot be determined at time of
object construction. For example, a bootstrap cluster does not exist when executing management
workflows, therefore a Kubeconfig isn't available to communicate with the cluster so must be passed
as contextual data.
*/
package workflowcontext
