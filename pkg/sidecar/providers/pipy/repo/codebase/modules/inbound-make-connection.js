pipy()

.import({
  __address: 'inbound-main',
})

.pipeline()
.branch(
  () => __address, (
    $=>$.connect(() => __address)
  ), (
    $=>$.chain()
  )
)